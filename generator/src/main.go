package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/sekerez/polka/generator/src/gospammer"
)

const (
	envPath  = "env/generator.env"
	transUrl = "/transaction"
	helloUrl = "/hello"
)

func main() {

	// Check for right args
	// args := os.Args
	// if len(args) < 2 {
	// 	log.Fatal("No transaction number provided.")
	// } else if len(args) > 4 {
	// 	log.Fatal("Too many arguments provided.")
	// }
	helloPtr := flag.Bool("h", false, "whether to send a hello GET request")
	workerPtr := flag.Uint("w", 3000, "the number of workers")
	transactionsPtr := flag.Uint("t", 100, "the number of transactions sent")
	flag.Parse()

	// Get url to api
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Environmental variables failed to load: %s\n", err)
	}
	mainUrl := os.Getenv("URL")

	// Say hello if asked!
	if *helloPtr {
		gospammer.SayHello(mainUrl + helloUrl)
	}

	log.Printf("Sending %d transactions with %d workers", *transactionsPtr, *workerPtr)
	badReqs := gospammer.TransactionSpammer(mainUrl+transUrl, *workerPtr, *transactionsPtr)
	log.Printf("Of all requests, %d were successful and %d failed.", *transactionsPtr-uint(badReqs), badReqs)
}
