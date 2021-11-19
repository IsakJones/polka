package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/IsakJones/polka/api/service/handlers"
)

const (
	transPath = "/transaction"
	helloPath = "/hello"
)

type Service struct {
	address  string
	listener net.Listener
	logger   *log.Logger
	mux      *http.ServeMux
	server   *http.Server
	ctx      context.Context
}

// New returns an uninitialized http service.
func New(address string, ctx context.Context) (serv *Service) {
	serv = &Service{
		ctx:     ctx,
		address: address,
		// date, time, and file
		logger: log.New(os.Stderr, "[main] ", log.LstdFlags),
	}
	return
}

// Start sets up a server and listener for incoming requests.
func (s *Service) Start() error {
	// Set up the mux
	s.mux = http.NewServeMux()
	s.mux.HandleFunc(transPath, handlers.Trans)
	s.mux.HandleFunc(helloPath, handlers.Hello)

	// Set up server
	s.server = &http.Server{
		Handler: s.mux,
	}

	// Set up listener
	lstn, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	s.listener = lstn

	// Activate server
	go func() {
		err := s.server.Serve(s.listener)
		if err != nil {
			s.logger.Fatalf("Error while serving: %s", err)
		}
	}()

	return nil
}

func (s *Service) Close() error {
	s.listener.Close()
	return nil
}

func (s *Service) Address() net.Addr {
	return s.listener.Addr()
}

func (s *Service) BaseContext() context.Context {
	return s.ctx
}
