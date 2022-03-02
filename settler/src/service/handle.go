package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sekerez/polka/settler/src/client"
	"github.com/sekerez/polka/settler/src/dbstore"
)

type settlementsManager struct {
	requested bool
	logger    *log.Logger
}

var cm settlementsManager

func initManager() {
	cm = settlementsManager{
		requested: false,
		logger:    log.New(os.Stderr, "[handler] ", log.LstdFlags|log.Lshortfile),
	}
}

// handle pings the cache once a request has arrived.
func handle(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	timeout, err := time.ParseDuration(r.FormValue("timeout"))
	if err == nil {
		// log.Printf("Detected timeout")
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	switch r.Method {
	case http.MethodPut:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case http.MethodDelete:
		w.WriteHeader(http.StatusMethodNotAllowed)

	case http.MethodGet:
		log.Printf("Got a snapshot request!")

		snapshot, err := client.RequestSnapshot()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error retrieving balances: %s", err)
			cm.logger.Printf("Error retrieving balances: %s", err)
			return
		}
		err = dbstore.InsertSnapshot(ctx, snapshot)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error sending snapshot to MongoDB database: %s", err)
			cm.logger.Printf("Error sending snapshot to MongoDB database: %s", err)
			return
		}
		_, err = w.Write(snapshot)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error writing snapshot to response: %s", err)
			cm.logger.Printf("Error writing snapshot to response: %s", err)
			return
		}
		cm.requested = true

	case http.MethodPost:
		// Make sure that the snapshot was requested
		if !cm.requested {
			err := errors.New("must request snapshot before requesting settlement").Error()
			cm.logger.Printf("Client requested snapshot before requesting settlement")
			http.Error(w, err, http.StatusBadRequest)
			return
		}

		// payload buffer might include data in the future
		payloadBuffer := new(bytes.Buffer)
		json.NewEncoder(payloadBuffer).Encode(struct{}{})

		// request settlement
		err := client.Settle(payloadBuffer)
		if err != nil {
			log.Fatalf("Error sending clearing request to cache: %s", err)
		}
		log.Printf("Successfully cleared balances.")
		fmt.Fprintf(w, "Successfully cleared balances.")
	}

}
