package types

import (
	"math/big"

	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/common/math"
)

// Message is a fully derived transaction and implements core.Message
//
// NOTE: In a future PR this will be removed.
type Message struct {
	to              *common.Address
	from            common.Address
	nonce           uint64
	amount          *big.Int
	gasLimit        uint64
	gasPrice        *big.Int
	data            []byte
	checkNonce      bool
	balanceTokenFee *big.Int
	gasFeeCap       *big.Int
	gasTipCap       *big.Int
}

func NewMessage(from common.Address, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, gasFeeCap *big.Int, gasTipCap *big.Int, data []byte, checkNonce bool, balanceTokenFee *big.Int, baseFee *big.Int) Message {
	if balanceTokenFee != nil {
		gasPrice = common.TRC21GasPrice
	}
	msg := Message{
		from:            from,
		to:              to,
		nonce:           nonce,
		amount:          amount,
		gasLimit:        gasLimit,
		gasFeeCap:       gasFeeCap,
		gasTipCap:       gasTipCap,
		gasPrice:        gasPrice,
		data:            data,
		checkNonce:      checkNonce,
		balanceTokenFee: balanceTokenFee,
	}
	if baseFee != nil {
		// If baseFee provided, set gasPrice to effectiveGasPrice.
		msg.gasPrice = math.BigMin(msg.gasPrice.Add(msg.gasTipCap, baseFee), msg.gasFeeCap)
	}
	return msg
}

func (m Message) From() common.Address      { return m.from }
func (m Message) BalanceTokenFee() *big.Int { return m.balanceTokenFee }
func (m Message) To() *common.Address       { return m.to }
func (m Message) GasPrice() *big.Int        { return m.gasPrice }
func (m Message) Value() *big.Int           { return m.amount }
func (m Message) Gas() uint64               { return m.gasLimit }
func (m Message) Nonce() uint64             { return m.nonce }
func (m Message) Data() []byte              { return m.data }
func (m Message) CheckNonce() bool          { return m.checkNonce }
func (m Message) GasFeeCap() *big.Int       { return m.gasFeeCap }
func (m Message) GasTipCap() *big.Int       { return m.gasTipCap }
