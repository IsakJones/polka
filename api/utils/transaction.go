package utils

type TransactionRecord interface {
	GetSenBank() string
	GetRecBank() string

	GetSenAcc() int
	GetRecAcc() int

	GetAmount() int
}

type Transaction struct {
	Sender   BankInfo
	Receiver BankInfo
	Amount   int
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
