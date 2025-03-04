package eip1559

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/core/types"
	"github.com/tomochain/tomochain/params"
)

// CalcBaseFee calculate base fee of block header
func CalcBaseFee(config *params.ChainConfig, parent *types.Header) *big.Int {
	// If the current block is the first EIP-1559 block, return the InitialBaseFee.
	if !config.IsEIP1559(parent.Number) {
		return new(big.Int).SetUint64(params.InitialBaseFee)
	}
	parentGasTarget := parent.GasLimit / config.ElasticityMultiplier()
	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if parent.GasUsed == parentGasTarget {
		return new(big.Int).Set(parent.BaseFee)
	}

	var (
		num   = new(big.Int)
		denom = new(big.Int)
	)
	if parent.GasUsed > parentGasTarget {
		num.SetUint64(parent.GasUsed - parentGasTarget)
		num.Mul(num, parent.BaseFee)
		num.Div(num, denom.SetUint64(parentGasTarget))
		num.Div(num, denom.SetUint64(config.BaseFeeChangeDenominator()))
		if num.Cmp(common.Big1) < 0 {
			return num.Add(parent.BaseFee, common.Big1)
		}
		return num.Add(parent.BaseFee, num)
	} else {
		num.SetUint64(parentGasTarget - parent.GasUsed)
		num.Mul(num, parent.BaseFee)
		num.Div(num, denom.SetUint64(parentGasTarget))
		num.Div(num, denom.SetUint64(config.BaseFeeChangeDenominator()))

		baseFee := num.Sub(parent.BaseFee, num)
		if baseFee.Cmp(common.Big0) < 0 {
			baseFee = common.Big0
		}
		return baseFee
	}
}

// VerifyEIP1559Header verifies some header attributes which were changed in EIP-1559,
// - gas limit check
// - basefee check
func VerifyEIP1559Header(config *params.ChainConfig, parent, header *types.Header) error {
	parentGasLimit := parent.GasLimit
	if !config.IsEIP1559(parent.Number) {
		parentGasLimit = parentGasLimit * config.ElasticityMultiplier()
	}
	if err := VerifyGasLimit(parentGasLimit, header.GasLimit); err != nil {
		return err
	}
	if header.BaseFee == nil {
		return errors.New("header is missing baseFee")
	}
	expectedBaseFee := CalcBaseFee(config, parent)
	if header.BaseFee.Cmp(expectedBaseFee) != 0 {
		return fmt.Errorf("invalid baseFee: have %s, want %s, parentBaseFee %s, parentGasUsed %d",
			header.BaseFee, expectedBaseFee, parent.BaseFee, parent.GasUsed)
	}
	return nil
}

// VerifyGaslimit verifies the header gas limit according increase/decrease
// in relation to the parent gas limit.
func VerifyGasLimit(parentGasLimit, headerGasLimit uint64) error {
	diff := int64(parentGasLimit) - int64(headerGasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parentGasLimit / params.GasLimitBoundDivisor
	if uint64(diff) >= limit {
		return fmt.Errorf("invalid gas limit: have %d, want %d +-= %d", headerGasLimit, parentGasLimit, limit-1)
	}
	if headerGasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit below %d", params.MinGasLimit)
	}
	return nil
}
