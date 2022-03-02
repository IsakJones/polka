package utils

import (
	"errors"
	"time"
)

const maxPayment = 100000

type Payment struct {
	Sender   BankInfo
	Receiver BankInfo
	Amount   int
	Time     time.Time
}

type BankInfo struct {
	Name    string
	Account int
}

func (ct *Payment) IsValidPayment() error {
	if ct.Amount > maxPayment {
		return errors.New("payments over $1000 are not allowed")
	}
	if ct.Amount < 0 {
		return errors.New("payment can't have a negative amount")
	}
	return nil
}
