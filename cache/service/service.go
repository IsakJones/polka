package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sekerez/polka/cache/memstore"
	"github.com/sekerez/polka/cache/utils"
)

const (
	path = "/balance"
)

// Service manages the main application functions.
type Service struct {
	logger   *log.Logger
	listener net.Listener
	server   *http.Server
	mux      *http.ServeMux
	ctx      context.Context
}

// New returns an uninitialized http service.
func New(u *url.URL, ctx context.Context) (*Service, error) {

	logger := log.New(os.Stderr, "[service] ", log.LstdFlags)
	port := fmt.Sprintf(":%s", u.Port())

	// Set up multiplexor
	mux := http.NewServeMux()
	mux.HandleFunc(path, balancesHandler)

	// Set up server
	server := &http.Server{
		Handler: mux,
		Addr:    port,
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
	if err != nil {
		return nil, err
	}
	// Set up listener
	listener, err := net.Listen("tcp", tcpAddr.String())
	if err != nil {
		return nil, err
	}

	// Successfully initialize service
	s := &Service{
		ctx:      ctx,
		mux:      mux,
		server:   server,
		listener: listener,
		logger:   logger,
	}

	return s, nil
}

// Start sets up a server and listener for incoming requests.
func (s *Service) Serve(errChan chan<- error) {
	errChan <- s.server.Serve(s.listener)
}

func balancesHandler(w http.ResponseWriter, req *http.Request) {

	var (
		ctx            context.Context
		cancel         context.CancelFunc
		currentBalance utils.SRBalance
	)

	// Spawn context with timeout if request has timeout
	timeout, err := time.ParseDuration(req.FormValue("Timeout"))
	if err == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	switch req.Method {
	case http.MethodPost:
		err := json.NewDecoder(req.Body).Decode(&currentBalance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		err = httpDo(
			ctx,
			&currentBalance,
			memstore.UpdateDues,
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

// httpDo calls the f function on the current transaction while abiding by the context.
func httpDo(ctx context.Context, cb *utils.SRBalance, f func(*utils.SRBalance) error) error {
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

func (s *Service) Close() (err error) {
	s.listener.Close()
	err = s.server.Shutdown(s.ctx)
	return
}
