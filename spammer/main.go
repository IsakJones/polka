package main

import (
	"log"
	"os"
	"strconv"

	"github.com/IsakJones/polka/spammer/gospammer"
)

const (
	mainUrl  = "http://localhost:8090"
	transUrl = "/transaction"
	helloUrl = "/hello"
)

func main() {

	// Check for right args
	args := os.Args
	if len(args) < 2 {
		log.Fatal("No transaction number provided.")
	} else if len(args) > 3 {
		log.Fatal("Too many arguments provided.")
	}

	// Parse number of requests
	requestNumber, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		log.Fatal("Argument must be an integer.")
	}
	if len(args) == 3 && args[2] == "hello" {
		gospammer.SayHello(mainUrl+helloUrl)
	}
	gospammer.TransactionSpammer(mainUrl+transUrl, int(requestNumber))
}
