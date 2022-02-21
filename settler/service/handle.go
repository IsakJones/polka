package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sekerez/polka/settler/client"
)

type clearingManager struct {
	requested bool
	logger    *log.Logger
}

const reqInterval = 10 * time.Second

var cm clearingManager

func initManager() {
	cm.requested = false
	cm.logger = log.New(os.Stderr, "[handler] ", log.LstdFlags|log.Lshortfile)

}

// handle pings the cache once a request has arrived.
func handle(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPut:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case http.MethodDelete:
		w.WriteHeader(http.StatusMethodNotAllowed)

	case http.MethodGet:
		payload, err := client.RequestClear()
		if err != nil {
			fmt.Fprintf(w, "Error retrieving balances: %s", err)
			cm.logger.Printf("Error retrieving balances: %s", err)
		}
		w.Write(payload)

	case http.MethodPost:
		// payload buffer might include data in the future
		payloadBuffer := new(bytes.Buffer)
		json.NewEncoder(payloadBuffer).Encode(struct{}{})
		err := client.ExecClear(payloadBuffer)
		if err != nil {
			log.Fatalf("Error sending clearing request to cache: %s", err)
		}
		log.Printf("Successfully cleared balances.")
		fmt.Fprintf(w, "Successfully cleared balances.")
	}

}
