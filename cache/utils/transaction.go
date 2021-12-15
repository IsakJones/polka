package utils

import "time"

// type Transaction struct {
// 	Sender   bankInfo
// 	Receiver bankInfo
// 	Amount   int
// 	Time     time.Time
// }

// type bankInfo struct {
// 	Name    string
// 	Account int
// }

// func (trans *Transaction) GetSenBank() string {
// 	return trans.Sender.Name
// }

// func (trans *Transaction) GetRecBank() string {
// 	return trans.Receiver.Name
// }

// func (trans *Transaction) GetSenAcc() int {
// 	return trans.Sender.Account
// }

// func (trans *Transaction) GetRecAcc() int {
// 	return trans.Receiver.Account
// }

// func (trans *Transaction) GetAmount() int {
// 	return trans.Amount
// }

// func (trans *Transaction) GetTime() time.Time {
// 	return trans.Time
// }

// func (trans *Transaction) SetSenBank(name string) {
// 	trans.Sender.Name = name
// }

// func (trans *Transaction) SetSenAcc(acc int) {
// 	trans.Sender.Account = acc
// }

// func (trans *Transaction) SetRecBank(name string) {
// 	trans.Receiver.Name = name
// }

// func (trans *Transaction) SetRecAcc(acc int) {
// 	trans.Sender.Account = acc
// }

// func (trans *Transaction) SetTime(time time.Time) {
// 	trans.Time = time
// }

// func (trans *Transaction) SetAmount(amount int) {
// 	trans.Amount = amount
// }

// Transaction is an interface for the Transaction struct in the transaction handler.
type Transaction interface {
	GetSenBank() string
	GetRecBank() string
	SetSenBank(string)
	SetRecBank(string)

	GetSenAcc() int
	GetRecAcc() int
	SetSenAcc(int)
	SetRecAcc(int)

	GetAmount() int
	SetAmount(int)
	GetTime() time.Time
	SetTime(time.Time)
}
