package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/IsakJones/polka/api/memstore"
	"github.com/IsakJones/polka/api/utils"
)

var counter uint64

// Trans handles http requests concerning transactions.
func Trans(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		// Read the Body
		var current utils.Transaction
		err := json.NewDecoder(req.Body).Decode(&current)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		atomic.AddUint64(&counter, 1)
		go memstore.UpdateDues(&current)
	}
}

func PrintProcessedTransactions() {
	fmt.Printf("Processed %d transactions by %s.\n", counter, time.Now().Format("15:04:05"))
}
