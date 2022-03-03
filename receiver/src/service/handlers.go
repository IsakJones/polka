package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sekerez/polka/receiver/src/client"
	"github.com/sekerez/polka/receiver/src/dbstore"
	"github.com/sekerez/polka/utils"
)

var validBanks = []string{
	"JP Morgan Chase",
	"Bank of America",
	"Wells Fargo",
	"Citigroup",
	"U.S. Bancorp",
	"Truist Financial",
	"PNC Financial Services Group",
	"TD Group US",
	"Bank of New York Mellon",
	"Capital One Financial",
}

// bankBalance stores information processed by the cache.
type bankBalance struct {
	Sender   utils.BankInfo
	Receiver utils.BankInfo
	Amount   int
}

var counter uint64

// handlePayment handles http requests concerning transactions.
func handlePayment(w http.ResponseWriter, req *http.Request) {

	// start := time.Now()
	// paymnt stands for current transaction
	var (
		ctx    context.Context
		cancel context.CancelFunc
		paymnt utils.Payment
	)

	// req.ParseForm()
	// for key, value := range req.Form {
	// 	log.Printf("%s: %s", key, value)
	// }
	// log.Printf("Passed form printing.")
	// Spawn context with timeout if request has timeout
	timeout, err := time.ParseDuration(req.FormValue("timeout"))
	if err == nil {
		// log.Printf("Detected timeout")
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Read body if post or delete request
	if req.Method == http.MethodPost || req.Method == http.MethodDelete {
		// Read the body
		// body, err := ioutil.ReadAll(req.Body)
		// if err != nil {
		// 	log.Printf("Error while reading body: %s", err.Error())
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	return
		// }
		err = json.NewDecoder(req.Body).Decode(&paymnt)
		if err != nil {
			log.Printf("Error decoding json: %s", err.Error())
			// log.Printf("Request body: %s", body)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = paymnt.IsValidPayment(); err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Multiplex according to method
	switch req.Method {
	case http.MethodPost:
		// Concurrently send request over to cache
		cacheErr := make(chan error)
		go sendTransactionToCache(&paymnt, paymnt.Amount, cacheErr)

		// innerStart := time.Now()
		// Insert transaction data into db
		err = dbstore.InsertPayment(ctx, &paymnt)
		if err != nil {
			log.Printf("Error with database: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// innerEnd := time.Now()
		// log.Printf("insert duration %s", innerEnd.Sub(innerStart))

		// check error from cache
		err = <-cacheErr
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Error from cache: %s", err)
		}
		atomic.AddUint64(&counter, 1)

	case http.MethodGet:
		// Scan table row in current transaction struct
		err = dbstore.GetPayment(ctx, &paymnt)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "%+v", paymnt)

		// Encode transaction data and send back to client
	// end := time.Now()
	// duration := end.Sub(start)
	// log.Printf("Handling request took %s", duration)
	case http.MethodPut:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case http.MethodDelete:

		// Concurrently update cache
		cacheErr := make(chan error)
		// Send the negative of the amount
		go sendTransactionToCache(&paymnt, -paymnt.Amount, cacheErr)

		// Update database
		err = dbstore.DeletePayment(ctx, &paymnt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// check error from cache
		err = <-cacheErr
		if err != nil {
			log.Printf("Error from cache: %s", err)
		}

	}
}

func sendTransactionToCache(paymnt *utils.Payment, amount int, cacheErr chan<- error) {
	payloadBuffer := new(bytes.Buffer)
	currentBalance := &bankBalance{
		Sender:   paymnt.Sender,
		Receiver: paymnt.Receiver,
		Amount:   amount,
	}
	json.NewEncoder(payloadBuffer).Encode(currentBalance)
	cacheErr <- client.SendTransactionUpdate(payloadBuffer)
}

func PrintProcessedTransactions() {
	log.Printf("Processed %d transactions.", counter)
}

// handleHello verifies that get requests work.
func handleHello(writer http.ResponseWriter, _ *http.Request) {

	fmt.Println("Successfully got a hello HTTP request!")
	fmt.Fprintf(writer, "Hello!!!\n")
}

// testHandler is used to monitor the machine's performance during long tasks
func testHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Second)
	return
}
