package eip1559

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/core/types"
	"github.com/tomochain/tomochain/params"
)

// copyConfig does a _shallow_ copy of a given config. Safe to set new values, but
// do not use e.g. SetInt() on the numbers. For testing only
func copyConfig(original *params.ChainConfig) *params.ChainConfig {
	return &params.ChainConfig{
		ChainId:                      &big.Int{},
		HomesteadBlock:               original.HomesteadBlock,
		DAOForkBlock:                 original.DAOForkBlock,
		DAOForkSupport:               original.DAOForkSupport,
		EIP150Block:                  original.EIP150Block,
		EIP150Hash:                   [32]byte{},
		EIP155Block:                  original.EIP155Block,
		EIP158Block:                  original.EIP158Block,
		ByzantiumBlock:               original.ByzantiumBlock,
		ConstantinopleBlock:          &big.Int{},
		TIP2019Block:                 &big.Int{},
		TIPSigningBlock:              &big.Int{},
		TIPRandomizeBlock:            &big.Int{},
		BlackListHFBlock:             &big.Int{},
		TIPTRC21FeeBlock:             &big.Int{},
		TIPTomoXBlock:                &big.Int{},
		TIPTomoXLendingBlock:         &big.Int{},
		TIPTomoXCancellationFeeBlock: &big.Int{},
		SaigonBlock:                  &big.Int{},
		EIP1559Block:                 original.EIP1559Block,
		Ethash:                       &params.EthashConfig{},
		Clique:                       &params.CliqueConfig{},
		Posv:                         &params.PosvConfig{},
	}
}

func config() *params.ChainConfig {
	config := copyConfig(params.TestChainConfig)
	config.EIP1559Block = big.NewInt(5)
	return config
}

// TestCalcBaseFee assumes all blocks are 1559-blocks
func TestCalcBaseFee(t *testing.T) {
	tests := []struct {
		parentBaseFee   int64
		parentGasLimit  uint64
		parentGasUsed   uint64
		expectedBaseFee int64
	}{
		{params.InitialBaseFee, 20000000, 10000000, params.InitialBaseFee}, // usage == target
		{params.InitialBaseFee, 20000000, 9000000, 987500000},              // usage below target
		{params.InitialBaseFee, 20000000, 11000000, 1012500000},            // usage above target
	}
	for i, test := range tests {
		parent := &types.Header{
			Number:   common.Big32,
			GasLimit: test.parentGasLimit,
			GasUsed:  test.parentGasUsed,
			BaseFee:  big.NewInt(test.parentBaseFee),
		}
		have := CalcBaseFee(config(), parent)
		want := big.NewInt(test.expectedBaseFee)
		fmt.Println("have:", have, "-", "want:", want)
		if have.Cmp(want) != 0 {
			t.Errorf("test %d: have %d  want %d, ", i, have, want)
		}
	}
	currentGasLimit := 84000000
	currentGasUsed := 25352
	currentBlock := &types.Header{
		Number:   common.Big32,
		GasLimit: uint64(currentGasLimit),
		GasUsed:  uint64(currentGasUsed),
		BaseFee:  big.NewInt(875075453),
	}
	expect := CalcBaseFee(config(), currentBlock)
	fmt.Println("expected", expect)
}

// TestBlockGasLimits tests the gasLimit checks for blocks both across
// the EIP-1559 boundary and post-1559 blocks
func TestBlockGasLimits(t *testing.T) {
	initial := new(big.Int).SetUint64(params.InitialBaseFee)

	for i, tc := range []struct {
		pGasLimit uint64
		pNum      int64
		gasLimit  uint64
		ok        bool
	}{
		// Transitions from non-EIP1559 to EIP1559
		{10000000, 4, 20000000, true},  // No change
		{10000000, 4, 20019530, true},  // Upper limit
		{10000000, 4, 20019531, false}, // Upper +1
		{10000000, 4, 19980470, true},  // Lower limit
		{10000000, 4, 19980469, false}, // Lower limit -1
		// EIP1559 to EIP1559
		{20000000, 5, 20000000, true},
		{20000000, 5, 20019530, true},  // Upper limit
		{20000000, 5, 20019531, false}, // Upper limit +1
		{20000000, 5, 19980470, true},  // Lower limit
		{20000000, 5, 19980469, false}, // Lower limit -1
		{40000000, 5, 40039061, true},  // Upper limit
		{40000000, 5, 40039062, false}, // Upper limit +1
		{40000000, 5, 39960939, true},  // lower limit
		{40000000, 5, 39960938, false}, // Lower limit -1
	} {
		parent := &types.Header{
			GasUsed:  tc.pGasLimit / 2,
			GasLimit: tc.pGasLimit,
			BaseFee:  initial,
			Number:   big.NewInt(tc.pNum),
		}
		header := &types.Header{
			GasUsed:  tc.gasLimit / 2,
			GasLimit: tc.gasLimit,
			BaseFee:  initial,
			Number:   big.NewInt(tc.pNum + 1),
		}
		err := VerifyEIP1559Header(config(), parent, header)
		if tc.ok && err != nil {
			t.Errorf("test %d: Expected valid header: %s", i, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("test %d: Expected invalid header", i)
		}
	}
}
