package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sekerez/polka/api/client"
	"github.com/sekerez/polka/api/dbstore"
)

type transaction struct {
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

func (trans *transaction) GetSenBank() string {
	return trans.Sender.Name
}

func (trans *transaction) GetRecBank() string {
	return trans.Receiver.Name
}

func (trans *transaction) GetSenAcc() int {
	return trans.Sender.Account
}

func (trans *transaction) GetRecAcc() int {
	return trans.Receiver.Account
}

func (trans *transaction) GetAmount() int {
	return trans.Amount
}

func (trans *transaction) GetTime() time.Time {
	return trans.Time
}

func (trans *transaction) SetSenBank(name string) {
	trans.Sender.Name = name
}

func (trans *transaction) SetSenAcc(accNum int) {
	trans.Sender.Account = accNum
}

func (trans *transaction) SetRecBank(name string) {
	trans.Receiver.Name = name
}

func (trans *transaction) SetRecAcc(accNum int) {
	trans.Receiver.Account = accNum
}

func (trans *transaction) SetTime(time time.Time) {
	trans.Time = time
}

func (trans *transaction) SetAmount(amount int) {
	trans.Amount = amount
}

var counter uint64

// Trans handles http requests concerning transactions.
func Transactions(w http.ResponseWriter, req *http.Request) {

	// start := time.Now()
	// ct stands for current transaction
	var (
		ctx    context.Context
		cancel context.CancelFunc
		ct     transaction
	)

	req.ParseForm()
	for key, value := range req.Form {
		log.Printf("%s: %s", key, value)
	}
	log.Printf("Passed form printing.")
	// Spawn context with timeout if request has timeout
	timeout, err := time.ParseDuration(req.FormValue("timeout"))
	if err == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Read body if post or delete request
	if req.Method == http.MethodPost || req.Method == http.MethodDelete {
		// Read the body
		err := json.NewDecoder(req.Body).Decode(&ct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	// Multiplex according to method
	switch req.Method {
	case http.MethodPost:

		// Concurrently send request over to cache
		cacheErr := make(chan error)
		go sendTransactionToCache(&ct, ct.Amount, cacheErr)

		// innerStart := time.Now()
		// Insert transaction data into db
		err = dbstore.InsertTransaction(ctx, &ct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// innerEnd := time.Now()
		// log.Printf("insert duration %s", innerEnd.Sub(innerStart))

		// check error from cache
		err = <-cacheErr
		if err != nil {
			log.Printf("Error from cache: %s", err)
		}
		atomic.AddUint64(&counter, 1)

	case http.MethodGet:
		// Scan table row in current transaction struct
		err = dbstore.GetTransaction(ctx, &ct)
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

		// Read the body
		err := json.NewDecoder(req.Body).Decode(&ct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		// Concurrently update cache
		cacheErr := make(chan error)
		// Send the negative of the amount
		go sendTransactionToCache(&ct, -ct.Amount, cacheErr)

		// Update database
		err = dbstore.DeleteTransaction(ctx, &ct)
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

func sendTransactionToCache(ct *transaction, amount int, cacheErr chan<- error) {
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
