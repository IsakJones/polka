package utils

import "time"

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
