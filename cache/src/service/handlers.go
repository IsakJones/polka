package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sekerez/polka/cache/src/memstore"
	"github.com/sekerez/polka/cache/src/utils"
)

func balancesHandler(w http.ResponseWriter, r *http.Request) {

	var (
		ctx            context.Context
		cancel         context.CancelFunc
		currentBalance utils.SRBalance
	)

	// Spawn context with timeout if request has timeout
	timeout, err := time.ParseDuration(r.FormValue("Timeout"))
	if err == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	switch r.Method {
	case http.MethodPost:
		err := json.NewDecoder(r.Body).Decode(&currentBalance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		err = enqueueBalance(
			ctx,
			&currentBalance,
			memstore.UpdateBalances,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	case http.MethodGet:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case http.MethodPut:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case http.MethodDelete:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func clearingHandler(w http.ResponseWriter, r *http.Request) {

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	// Spawn context with timeout if request has timeout
	timeout, err := time.ParseDuration(r.FormValue("Timeout"))
	if err == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		// The only thing you need to do is take the snapshot and send it back
		balances, err := enqueueSnapRequest(ctx, memstore.GetSnapshot)
		if balances == nil || err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(balances)

	case http.MethodPost:
		err := memstore.SettleSnapshot()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	case http.MethodPut:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case http.MethodDelete:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// enqueueBalance calls the f function on the current transaction while abiding by the context.
func enqueueBalance(ctx context.Context, cb *utils.SRBalance, f func(*utils.SRBalance) error) error {
	// Update the dues in a goroutine and pass the result to fChan
	fChan := make(chan error)
	go func() { fChan <- f(cb) }()

	// Return an error if the context times out or if the function returns an error.
	select {
	case <-ctx.Done():
		<-fChan
		return ctx.Err()
	case err := <-fChan:
		return err
	}
}

// enqueueBalance calls the f function on the current transaction while abiding by the context.
func enqueueSnapRequest(ctx context.Context, f func() (*memstore.Snapshot, error)) (*memstore.Snapshot, error) {
	// Make error channel
	errChan := make(chan error)
	snapChan := make(chan *memstore.Snapshot)

	// Pass function result to channel
	go func() {
		snap, err := f()
		errChan <- err
		snapChan <- snap
	}()

	// Return an error if the context times out or if the function returns an error.
	select {
	case <-ctx.Done():
		<-errChan
		<-snapChan
		memstore.CancelSnapshot()
		return nil, ctx.Err()
	case err := <-errChan:
		snap := <-snapChan
		return snap, err
	}
}
