package utils

import "time"

type TransactionRecord interface {
	GetSenBank() string
	GetRecBank() string

	GetSenAcc() int
	GetRecAcc() int

	GetAmount() int
	GetTime() time.Time
}

type Transaction struct {
	Sender   BankInfo
	Receiver BankInfo
	Amount   int
	Time time.Time
}

type BankInfo struct {
	Name string
	Account int
}

func (trans *Transaction) GetSenBank() string {
	return trans.Sender.Name
}

func (trans *Transaction) GetRecBank() string {
	return trans.Receiver.Name
}

func (trans *Transaction) GetSenAcc() int {
	return trans.Sender.Account
}

func (trans *Transaction) GetRecAcc() int {
	return trans.Receiver.Account
}

func (trans *Transaction) GetAmount() int {
	return trans.Amount
}

func (trans *Transaction) GetTime() time.Time {
	return trans.Time
}