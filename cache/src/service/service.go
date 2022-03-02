package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
)

const (
	balancePath  = "/balance"
	clearingPath = "/settle"
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
	mux.HandleFunc(balancePath, balancesHandler)
	mux.HandleFunc(clearingPath, clearingHandler)

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

func (s *Service) Close() (err error) {
	s.listener.Close()
	err = s.server.Shutdown(s.ctx)
	return
}
