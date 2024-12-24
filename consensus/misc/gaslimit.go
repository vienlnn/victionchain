package misc

import (
	"fmt"

	"github.com/tomochain/tomochain/params"
)

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
