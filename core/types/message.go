package types

// Message is a fully derived transaction and implements core.Message
//
// NOTE: In a future PR this will be removed.
// type Message struct {
// 	to              *common.Address
// 	from            common.Address
// 	nonce           uint64
// 	amount          *big.Int
// 	gasLimit        uint64
// 	gasPrice        *big.Int
// 	data            []byte
// 	checkNonce      bool
// 	balanceTokenFee *big.Int
// 	gasFeeCap       *big.Int
// 	gasTipCap       *big.Int
// }

// func NewMessage(from common.Address, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool, balanceTokenFee *big.Int) Message {
// 	if balanceTokenFee != nil {
// 		gasPrice = common.TRC21GasPrice
// 	}
// 	return Message{
// 		from:            from,
// 		to:              to,
// 		nonce:           nonce,
// 		amount:          amount,
// 		gasLimit:        gasLimit,
// 		gasPrice:        gasPrice,
// 		data:            data,
// 		checkNonce:      checkNonce,
// 		balanceTokenFee: balanceTokenFee,
// 	}
// }

// func (m Message) From() common.Address      { return m.from }
// func (m Message) BalanceTokenFee() *big.Int { return m.balanceTokenFee }
// func (m Message) To() *common.Address       { return m.to }
// func (m Message) GasPrice() *big.Int        { return m.gasPrice }
// func (m Message) Value() *big.Int           { return m.amount }
// func (m Message) Gas() uint64               { return m.gasLimit }
// func (m Message) Nonce() uint64             { return m.nonce }
// func (m Message) Data() []byte              { return m.data }
// func (m Message) CheckNonce() bool          { return m.checkNonce }
// func (m Message) GasFeeCap() *big.Int       { return m.gasFeeCap }
// func (m Message) GasTipCap() *big.Int       { return m.gasTipCap }
