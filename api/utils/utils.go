package utils

type Transaction struct {
	Sender   string
	Receiver string
	Amount   int
}

func (trans *Transaction) GetSender() string {
	return trans.Sender
}

func (trans *Transaction) GetReceiver() string {
	return trans.Receiver
}

func (trans *Transaction) GetAmount() int {
	return trans.Amount
}

type Trans interface {
	GetSender() string
	GetReceiver() string
	GetAmount() int
}

func SumMap64(mp map[string]int64) (result int64) {
	for _, value := range mp {
		result += value
	}
	return
}
