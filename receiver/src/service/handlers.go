package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sekerez/polka/receiver/src/client"
	"github.com/sekerez/polka/receiver/src/dbstore"
)

const maxPayment = 100000

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

type payment struct {
	Sender   bankInfo
	Receiver bankInfo
	Amount   int
	Time     time.Time
}

type bankInfo struct {
	Name    string
	Account int
}

// bankBalance stores information processed by the cache.
type bankBalance struct {
	Sender   bankInfo
	Receiver bankInfo
	Amount   int
}

func (ct *payment) isValidPayment() error {
	if ct.Amount > maxPayment {
		return errors.New("payments over $1000 are not allowed")
	}
	if ct.Amount < 0 {
		return errors.New("payment can't have a negative amount")
	}
	return nil
}

func (trans *payment) GetSenBank() string {
	return trans.Sender.Name
}

func (trans *payment) GetRecBank() string {
	return trans.Receiver.Name
}

func (trans *payment) GetSenAcc() int {
	return trans.Sender.Account
}

func (trans *payment) GetRecAcc() int {
	return trans.Receiver.Account
}

func (trans *payment) GetAmount() int {
	return trans.Amount
}

func (trans *payment) GetTime() time.Time {
	return trans.Time
}

func (trans *payment) SetSenBank(name string) {
	trans.Sender.Name = name
}

func (trans *payment) SetSenAcc(accNum int) {
	trans.Sender.Account = accNum
}

func (trans *payment) SetRecBank(name string) {
	trans.Receiver.Name = name
}

func (trans *payment) SetRecAcc(accNum int) {
	trans.Receiver.Account = accNum
}

func (trans *payment) SetTime(time time.Time) {
	trans.Time = time
}

func (trans *payment) SetAmount(amount int) {
	trans.Amount = amount
}

var counter uint64

// handlePayment handles http requests concerning transactions.
func handlePayment(w http.ResponseWriter, req *http.Request) {

	// start := time.Now()
	// ct stands for current transaction
	var (
		ctx    context.Context
		cancel context.CancelFunc
		ct     payment
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
		err = json.NewDecoder(req.Body).Decode(&ct)
		if err != nil {
			log.Printf("Error decoding json: %s", err.Error())
			// log.Printf("Request body: %s", body)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// if err = ct.isValidPayment(); err != nil {
		// 	log.Printf("%s", err.Error())
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	return
		// }
	}

	// Multiplex according to method
	switch req.Method {
	case http.MethodPost:
		// Concurrently send request over to cache
		cacheErr := make(chan error)
		go sendTransactionToCache(&ct, ct.Amount, cacheErr)

		// innerStart := time.Now()
		// Insert transaction data into db
		err = dbstore.InsertPayment(ctx, &ct)
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
		err = dbstore.GetPayment(ctx, &ct)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "%+v", ct)

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
		go sendTransactionToCache(&ct, -ct.Amount, cacheErr)

		// Update database
		err = dbstore.DeletePayment(ctx, &ct)
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

func sendTransactionToCache(ct *payment, amount int, cacheErr chan<- error) {
	payloadBuffer := new(bytes.Buffer)
	currentBalance := &bankBalance{
		Sender:   ct.Sender,
		Receiver: ct.Receiver,
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
