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
	paymentView = "/payment"
	helloView   = "/hello"
)

// Service manages the main application functions.
type Service struct {
	logger   *log.Logger
	listener net.Listener
	server   *http.Server
	mux      *http.ServeMux
	ctx      context.Context
}

func (s *Service) Address() net.Addr {
	return s.listener.Addr()
}

// New returns an uninitialized http service.
func New(u *url.URL, ctx context.Context) (*Service, error) {

	logger := log.New(os.Stderr, "[service] ", log.LstdFlags|log.Lshortfile)

	// Format port
	port := fmt.Sprintf(":%s", u.Port())

	tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
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
	mux.HandleFunc(paymentView, handlePayment)
	mux.HandleFunc(helloView, handleHello)

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

	return s, nil
}

// Start sets up a server and listener for incoming requests.
func (s *Service) Serve(errChan chan<- error) {
	s.logger.Printf("Listening on %s", s.server.Addr)

	errChan <- s.server.Serve(s.listener)
}

func (s *Service) Close() (err error) {
	s.listener.Close()
	err = s.server.Shutdown(s.ctx)
	return
}
