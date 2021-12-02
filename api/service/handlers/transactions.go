package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/IsakJones/polka/api/memstore"
	// "github.com/IsakJones/polka/api/dbstore"
	// "github.com/IsakJones/polka/api/utils"
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
func Trans(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		// Read the Body
		var current transaction
		err := json.NewDecoder(req.Body).Decode(&current)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		atomic.AddUint64(&counter, 1)
		go memstore.UpdateDues(&current)
		// TODO blocker or other way
	}
}

func PrintProcessedTransactions() {
	fmt.Printf("Processed %d transactions by %s.\n", counter, time.Now().Format("15:04:05"))
}
