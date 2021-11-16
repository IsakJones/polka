package main

import "github.com/IsakJones/polka/spammer/gospammer"


func main() {
	transactionDest := "http://localhost:8090/transaction"
	helloDest := "http://localhost:8090/hello"

	gospammer.SayHello(helloDest)
	gospammer.SendTransaction(transactionDest)
}