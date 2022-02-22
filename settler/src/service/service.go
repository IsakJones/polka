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
	path = "/settle"
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

	logger := log.New(os.Stderr, "[service] ", log.LstdFlags|log.Lshortfile)

	// Configure TCP connection
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%s", u.Port()))
	if err != nil {
		return nil, err
	}

	// Set up listener
	listener, err := net.Listen("tcp", tcpAddr.String())
	if err != nil {
		return nil, err
	}

	// Set up multiplexor
	mux := http.NewServeMux()
	mux.HandleFunc(path, handle)

	// Set up server
	server := &http.Server{
		Handler: mux,
	}

	// Successfully initialize service
	s := &Service{
		ctx:      ctx,
		mux:      mux,
		logger:   logger,
		server:   server,
		listener: listener,
	}

	// Initialize settlements manager
	initManager()

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
