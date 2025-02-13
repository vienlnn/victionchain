package gasprice

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"sync/atomic"

	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/core/types"
	"github.com/tomochain/tomochain/log"
	"github.com/tomochain/tomochain/rpc"
)

var (
	errInvalidPercentile = errors.New("invalid reward percentile")
	errRequestBeyondHead = errors.New("request beyond head block")
)

const (
	// maxBlockFetchers is the max number of goroutines to spin up to pull blocks
	// for the fee history calculation (mostly relevant for LES).
	maxBlockFetchers = 4
	// maxQueryLimit is the max number of requested percentiles.
	maxQueryLimit    = 100
	maxHeaderHistory = 1024
	maxBlockHistory  = 1024
)

// blockFees represents a single block for processing
type blockFees struct {
	// set by the caller
	blockNumber uint64
	header      *types.Header
	block       *types.Block // only set if reward percentiles are requested
	receipts    types.Receipts
	// filled by processBlock
	results processedFees
	err     error
}

type cacheKey struct {
	number      uint64
	percentiles string
}

// processedFees contains the results of a processed block.
type processedFees struct {
	reward                       []*big.Int
	baseFee, nextBaseFee         *big.Int
	gasUsedRatio                 float64
	blobGasUsedRatio             float64
	blobBaseFee, nextBlobBaseFee *big.Int
}

// txGasAndReward is sorted in ascending order based on reward
type txGasAndReward struct {
	gasUsed uint64
	reward  *big.Int
}

func (oracle *Oracle) processBlock(bf *blockFees, percentiles []float64) {
	// before [eip-1559]
	bf.results.baseFee = new(big.Int)
	bf.results.nextBaseFee = new(big.Int)

	// before [eip-4884]
	bf.results.blobBaseFee = new(big.Int)
	bf.results.nextBlobBaseFee = new(big.Int)

	// calc gasUsed ratio for normal gas => gasUsed/gasLimit
	bf.results.gasUsedRatio = float64(bf.header.GasUsed) / float64(bf.header.GasLimit)

	// reward percentiles not request => null
	if len(percentiles) == 0 {
		return
	}

	//fmt.Println("process:", bf.blockNumber, percentiles)
	//fmt.Println("process:receipt", bf.blockNumber, bf.receipts, bf.block.Transactions())
	//fmt.Println("process:condition", bf.blockNumber, bf.block == nil, bf.receipts == nil, len(bf.block.Transactions()), (bf.receipts == nil || len(bf.block.Transactions()) != 0))

	if bf.block == nil || (bf.receipts == nil && len(bf.block.Transactions()) != 0) {
		log.Error("Block or receipts are missing while reward percentiles are requested")
		return
	}

	bf.results.reward = make([]*big.Int, len(percentiles))

	if len(bf.block.Transactions()) == 0 {
		// return an all zero row if there are no transactions to gather data from
		for i := range bf.results.reward {
			bf.results.reward[i] = new(big.Int)
		}
		return
	}

	sorter := make([]txGasAndReward, len(bf.block.Transactions()))
	for i, tx := range bf.block.Transactions() {
		reward := big.NewInt(int64(tx.Gas()))
		sorter[i] = txGasAndReward{gasUsed: bf.receipts[i].GasUsed, reward: reward}
	}
	slices.SortStableFunc(sorter, func(a txGasAndReward, b txGasAndReward) int {
		return a.reward.Cmp(b.reward)
	})

	var txIndex int
	sumGasUsed := sorter[0].gasUsed

	for i, p := range percentiles {
		thresholdGasUsed := uint64(float64(bf.block.GasUsed()) * p / 100)
		for sumGasUsed < thresholdGasUsed && txIndex < len(bf.block.Transactions())-1 {
			txIndex++
			sumGasUsed += sorter[txIndex].gasUsed
		}
		bf.results.reward[i] = sorter[txIndex].reward
	}
}

func (oracle *Oracle) resolveBlockRange(ctx context.Context, reqEnd rpc.BlockNumber, blocks uint64) (*types.Block, []*types.Receipt, uint64, uint64, error) {
	var (
		headBlock       *types.Header
		pendingBlock    *types.Block
		pendingReceipts types.Receipts
		err             error
	)

	if headBlock, err = oracle.backend.HeaderByNumber(ctx, rpc.LatestBlockNumber); err != nil {
		return nil, nil, 0, 0, err
	}

	head := rpc.BlockNumber(headBlock.Number.Uint64())

	// Fail if request block is beyond the chain's current head.
	if head < reqEnd {
		return nil, nil, 0, 0, fmt.Errorf("%w: requested %d, head %d", errRequestBeyondHead, reqEnd, head)
	}

	if reqEnd < 0 {
		var (
			resolved *types.Header
			err      error
		)

		switch reqEnd {
		case rpc.PendingBlockNumber:
			pendingBlock, _ = oracle.backend.Pending()
			if pendingBlock != nil {
				resolved = pendingBlock.Header()
			} else {
				resolved = headBlock
				blocks--
			}
		case rpc.LatestBlockNumber:
			resolved = headBlock
		case rpc.FinalizedBlockNumber:
			finalized, fErr := oracle.backend.BlockByNumber(ctx, rpc.FinalizedBlockNumber)
			if fErr != nil {
				err = fErr
			}
			resolved = finalized.Header()
		}
		if resolved == nil || err != nil {
			return nil, nil, 0, 0, err
		}
		// Absolute number resolved.
		reqEnd = rpc.BlockNumber(resolved.Number.Uint64())
	}

	// If there are no blocks to return, short circuit.
	if blocks == 0 {
		return nil, nil, 0, 0, nil
	}
	// Ensure not trying to retrieve before genesis.
	if uint64(reqEnd+1) < blocks {
		blocks = uint64(reqEnd + 1)
	}
	return pendingBlock, pendingReceipts, uint64(reqEnd), blocks, nil
}

func (oracle *Oracle) FeeHistory(ctx context.Context, blocks uint64, unresolvedLastBlock rpc.BlockNumber, rewardPercentiles []float64) (*big.Int, [][]*big.Int, []*big.Int, []float64, []*big.Int, []float64, error) {
	if blocks < 1 {
		return common.Big0, nil, nil, nil, nil, nil, nil // returning with no data and no error means there are no retrievable blocks
	}

	maxFeeHistory := maxHeaderHistory
	if len(rewardPercentiles) != 0 {
		maxFeeHistory = maxBlockHistory
	}

	if len(rewardPercentiles) > maxQueryLimit {
		return common.Big0, nil, nil, nil, nil, nil, fmt.Errorf("%w: over the query limit %d", errInvalidPercentile, maxQueryLimit)
	}

	if blocks > uint64(maxFeeHistory) {
		log.Warn("Sanitizing fee history length", "requested", blocks, "truncated", maxFeeHistory)
		blocks = uint64(maxFeeHistory)
	}

	for i, p := range rewardPercentiles {
		if p < 0 || p > 100 {
			return common.Big0, nil, nil, nil, nil, nil, fmt.Errorf("%w: %f", errInvalidPercentile, p)
		}
		if i > 0 && p <= rewardPercentiles[i-1] {
			return common.Big0, nil, nil, nil, nil, nil, fmt.Errorf("%w: #%d:%f >= #%d:%f", errInvalidPercentile, i-1, rewardPercentiles[i-1], i, p)
		}
	}

	var (
		pendingBlock    *types.Block
		pendingReceipts []*types.Receipt
		err             error
	)
	pendingBlock, pendingReceipts, lastBlock, blocks, err := oracle.resolveBlockRange(ctx, unresolvedLastBlock, blocks)
	if err != nil || blocks == 0 {
		return common.Big0, nil, nil, nil, nil, nil, err
	}
	oldestBlock := lastBlock + 1 - blocks

	var next atomic.Uint64
	next.Store(oldestBlock)
	results := make(chan *blockFees, blocks)

	percentileKey := make([]byte, 8*len(rewardPercentiles))
	for i, p := range rewardPercentiles {
		binary.LittleEndian.PutUint64(percentileKey[i*8:(i+1)*8], math.Float64bits(p))
	}

	for i := 0; i < maxBlockFetchers && i < int(blocks); i++ {
		go func() {
			// Retrieve the next block number to fetch with this goroutine
			for {
				blockNumber := next.Add(1) - 1
				if blockNumber > lastBlock {
					return
				}
				fees := &blockFees{blockNumber: blockNumber}
				if pendingBlock != nil && blockNumber > pendingBlock.NumberU64() {
					fees.block, fees.receipts = pendingBlock, pendingReceipts
					fees.header = fees.block.Header()
					oracle.processBlock(fees, rewardPercentiles)
					results <- fees
				} else {
					if len(rewardPercentiles) != 0 {
						fees.block, fees.err = oracle.backend.BlockByNumber(ctx, rpc.BlockNumber(blockNumber))
						if fees.block != nil && fees.err == nil {
							fees.receipts, fees.err = oracle.backend.GetReceipts(ctx, fees.block.Hash())
							// fmt.Println("block", fees.block.Number(), fees.receipts)
							fees.header = fees.block.Header()
						}
					} else {
						fees.header, fees.err = oracle.backend.HeaderByNumber(ctx, rpc.BlockNumber(blockNumber))
					}
					if fees.header != nil && fees.err == nil {
						oracle.processBlock(fees, rewardPercentiles)
					}
					results <- fees
				}
			}
		}()
	}
	var (
		reward           = make([][]*big.Int, blocks)
		baseFee          = make([]*big.Int, blocks+1)
		gasUsedRatio     = make([]float64, blocks)
		blobGasUsedRatio = make([]float64, blocks)
		blobBaseFee      = make([]*big.Int, blocks+1)
		firstMissing     = blocks
	)

	for ; blocks > 0; blocks-- {
		fees := <-results
		if fees.err != nil {
			return common.Big0, nil, nil, nil, nil, nil, fees.err
		}
		i := fees.blockNumber - oldestBlock
		if fees.results.baseFee != nil {
			reward[i], baseFee[i], baseFee[i+1], gasUsedRatio[i] = fees.results.reward, fees.results.baseFee, fees.results.nextBaseFee, fees.results.gasUsedRatio
			blobGasUsedRatio[i], blobBaseFee[i], blobBaseFee[i+1] = fees.results.blobGasUsedRatio, fees.results.blobBaseFee, fees.results.nextBlobBaseFee
		} else {
			// getting no block and no error means we are requesting into the future (might happen because of a reorg)
			if i < firstMissing {
				firstMissing = i
			}
		}

	}
	if firstMissing == 0 {
		return common.Big0, nil, nil, nil, nil, nil, nil
	}

	if len(rewardPercentiles) != 0 {
		reward = reward[:firstMissing]
	} else {
		reward = nil
	}

	baseFee, gasUsedRatio = baseFee[:firstMissing+1], gasUsedRatio[:firstMissing]
	blobBaseFee, blobGasUsedRatio = blobBaseFee[:firstMissing+1], blobGasUsedRatio[:firstMissing]
	return new(big.Int).SetUint64(oldestBlock), reward, baseFee, gasUsedRatio, blobBaseFee, blobGasUsedRatio, nil
}
