package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/sekerez/polka/generator/src/gospammer"
)

const (
	envPath = "env/generator.env"
)

func main() {

	// Check if we should send a hello!
	helloPtr := flag.Bool("h", false, "whether to send a hello GET request")

	// Check if we should spam transactions, and if so how many
	workerPtr := flag.Uint("w", 3000, "the number of workers")
	transactionsPtr := flag.Uint("t", 100, "the number of transactions sent")

	// Check if we should take a snapshot
	getSnapshotPtr := flag.Bool("gs", false, "whether to get a snapshot")
	settleBalancesPtr := flag.Bool("sb", false, "whether to settle balances given a snapshot")

	// Parse flags
	flag.Parse()

	// Get url to services
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Environmental variables failed to load: %s\n", err)
	}

	mainDest := os.Getenv("MAINURL")
	helloDest := os.Getenv("HELLOURL")
	settleDest := os.Getenv("SETTLERURL")

	if *getSnapshotPtr {
		_, err := getSnapshot(settleDest)
		if err != nil {
			log.Fatalf("Error requesting snapshot: %s", err.Error())
		}
		return
	}

	if *settleBalancesPtr {
		err := settleBalances(settleDest)
		if err != nil {
			log.Printf("Error requesting snapshot: %s", err.Error())
			return
		}
		log.Printf("Balances have been settled successfully.")
		return
	}

	// Say hello if asked!
	if *helloPtr {
		gospammer.SayHello(helloDest)
	}

	log.Printf("Sending %d transactions with %d workers", *transactionsPtr, *workerPtr)
	badReqs := gospammer.TransactionSpammer(mainDest, *workerPtr, *transactionsPtr)
	log.Printf("Of all requests, %d were successful and %d failed.", *transactionsPtr-uint(badReqs), badReqs)
}
