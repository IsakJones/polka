package utils

// BankBalance transfers bank balance data from the cache to the database.
type BankBalance struct {
	Name    string
	Balance int64
}

// AccountBalance transfers account balance data from the cache to the database.
type AccountBalance struct {
	Num     uint16
	Balance int
}

// SRBalance captures data from the api and feeds it into the cache.
type SRBalance struct {
	Sender   *bankInfo
	Receiver *bankInfo
	Amount   int
}

type bankInfo struct {
	Name    string
	Account uint16
}
