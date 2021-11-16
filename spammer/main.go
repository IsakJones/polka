package main

import "github.com/IsakJones/polka/spammer/gospammer"

const (
	requestNumber = 2000
	transactionDest = "http://localhost:8090/transaction"
	// helloDest = "http://localhost:8090/hello"
)

func main() {

	// gospammer.SayHello(helloDest)
	// gospammer.SendTransaction(transactionDest)

	gospammer.TransactionSpammer(transactionDest, requestNumber)
}
