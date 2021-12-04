package utils

import "time"

// Transaction is an interface for the Transaction struct in the transaction handler.
type Transaction interface {
	GetSenBank() string
	GetRecBank() string

	GetSenAcc() int
	GetRecAcc() int

	GetAmount() int
	GetTime() time.Time
}
