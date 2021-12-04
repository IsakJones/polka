package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/IsakJones/polka/api/dbstore"
	"github.com/IsakJones/polka/api/memstore"
	"github.com/IsakJones/polka/api/utils"
	// "github.com/IsakJones/polka/api/dbstore"
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

var counter uint64

// Trans handles http requests concerning transactions.
func Transactions(writer http.ResponseWriter, req *http.Request) {

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	// Spawn context with timeout if request has timeout
	timeout, err := time.ParseDuration(req.FormValue("timeout"))
	if err == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		// log.Printf("Creating context without timeout.")
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Multiplex according to method
	switch req.Method {
	case http.MethodPost:
		// Read the Body
		var currentTransaction transaction
		err := json.NewDecoder(req.Body).Decode(&currentTransaction)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		// Insert transaction data into db
		// TODO wrap in httpDo, or replace memstore call with that
		err = dbstore.InsertTransaction(ctx, &currentTransaction)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		// Update dues with context
		httpDo(
			ctx,
			&currentTransaction,
			memstore.UpdateDues,
		)
		atomic.AddUint64(&counter, 1)
	}
}

// httpDo calls the f function on the current transaction while abiding by the context.
func httpDo(ctx context.Context, ct utils.Transaction, f func(utils.Transaction) error) error {
	// Update the dues in a goroutine and pass the result to fChan
	fChan := make(chan error)
	go func() { fChan <- f(ct) }()

	// Return an error if the context times out or if the function returns an error.
	select {
	case <-ctx.Done():
		<-fChan
		return ctx.Err()
	case err := <-fChan:
		return err
	}
}

func PrintProcessedTransactions() {
	fmt.Printf("Processed %d transactions by %s.\n", counter, time.Now().Format("15:04:05"))
}
