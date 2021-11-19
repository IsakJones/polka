package main

import (
	"log"
	"os"
	"strconv"
	
	"github.com/IsakJones/polka/spammer/gospammer"
)

const (
	requestNumber   = 9600
	transactionDest = "http://localhost:8090/transaction"
	// helloDest = "http://localhost:8090/hello"
)

func main() {

	// Check for right args
	args := os.Args
	if len(args) < 2 {
		log.Fatal("No transaction number provided.")
	} else if len(args) > 2 {
		log.Fatal("Too many arguments provided.")
	}

	// Parse number of requests 
	requestNumber, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		log.Fatal("Argument must be an integer.")
	}
	gospammer.TransactionSpammer(transactionDest, int(requestNumber))
}
