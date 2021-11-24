package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/IsakJones/polka/spammer/gospammer"
)

const (
	envPath  = "spammer.env"
	transUrl = "/transaction"
	helloUrl = "/hello"
)

func main() {

	// Get url to api
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Environmental variables failed to load: %s\n", err)
	}
	mainUrl := os.Getenv("URL")

	// Check for right args
	args := os.Args
	if len(args) < 2 {
		log.Fatal("No transaction number provided.")
	} else if len(args) > 3 {
		log.Fatal("Too many arguments provided.")
	}

	// Parse number of requests
	requestNumber, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal("Argument must be an integer.")
	}
	// Say hello if asked!
	if len(args) == 3 && args[2] == "hello" {
		gospammer.SayHello(mainUrl + helloUrl)
	}

	gospammer.TransactionSpammer(mainUrl+transUrl, int(requestNumber))
}
