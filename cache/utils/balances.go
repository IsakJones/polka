package utils

// BankBalance transfers bank balance data from the cache to the database.
type BankBalance struct {
	Name    string
	Balance int64
}

// SRBalance captures data from the api and feeds it into the cache.
type SRBalance struct {
	Sender   string
	Receiver string
	Amount   int
}
