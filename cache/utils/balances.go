package utils

// Balance transfers data to update the database accounts ledger.
type Balance struct {
	Bank    string
	Account uint16
	Balance int
}

// BankBalance transfers bank balance data from the cache to the database.
type BankBalance struct {
	Name    string
	Balance int64
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
