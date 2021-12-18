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
	var (
		ctx                context.Context
		cancel             context.CancelFunc
		currentTransaction transaction
	)

	// Spawn context with timeout if request has timeout
	timeout, err := time.ParseDuration(req.FormValue("Timeout"))
	if err == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Multiplex according to method
	switch req.Method {
	case http.MethodPost:

		// Read the Body
		err := json.NewDecoder(req.Body).Decode(&currentTransaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Concurrently send request over to cache
		cacheErr := make(chan error)

		go func() {
			payloadBuffer := new(bytes.Buffer)
			currentBalance := &bankBalance{
				Sender:   currentTransaction.Sender,
				Receiver: currentTransaction.Receiver,
				Amount:   currentTransaction.Amount,
			}
			json.NewEncoder(payloadBuffer).Encode(currentBalance)
			cacheErr <- client.SendTransactionUpdate(payloadBuffer)
		}()
		// innerStart := time.Now()
		// Insert transaction data into db
		err = dbstore.InsertTransaction(ctx, &currentTransaction)
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
		err = dbstore.GetTransaction(ctx, &currentTransaction)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "%+v", currentTransaction)

		// Encode transaction data and send back to client
	// end := time.Now()
	// duration := end.Sub(start)
	// log.Printf("Handling request took %s", duration)
	case http.MethodPut:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case http.MethodDelete:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func PrintProcessedTransactions() {
	log.Printf("Processed %d transactions.", counter)
}
