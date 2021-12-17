package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/sekerez/polka/api/service/handlers"
	"github.com/sekerez/polka/api/utils"
)

const (
	transPath = "/transaction"
	helloPath = "/hello"
)

// Service manages the main application functions.
type Service struct {
	logger   *log.Logger
	conf     utils.Config
	listener net.Listener
	server   *http.Server
	mux      *http.ServeMux
	wg       sync.WaitGroup
	ctx      context.Context
}

func (s *Service) Address() net.Addr {
	return s.listener.Addr()
}

func (s *Service) BaseContext() context.Context {
	return s.ctx
}

// New returns an uninitialized http service.
func New(conf utils.Config, ctx context.Context) (*Service, error) {

	var wg sync.WaitGroup

	tcpAddr, err := net.ResolveTCPAddr("tcp4", conf.GetListenPort())
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
	mux.HandleFunc(transPath, handlers.Transactions)
	mux.HandleFunc(helloPath, handlers.Hello)

	// Set up server
	server := &http.Server{
		Handler: mux,
	}

	// Successfully initialize service
	s := &Service{
		wg:       wg,
		ctx:      ctx,
		mux:      mux,
		conf:     conf,
		server:   server,
		listener: listener,
		logger:   log.New(os.Stderr, "[main] ", log.LstdFlags),
	}

	// Start listening for requests
	s.wg.Add(1)
	go s.Start()

	return s, nil
}

// Start sets up a server and listener for incoming requests.
func (s *Service) Start() {
	defer s.wg.Done()

	// Activate server
	err := s.server.Serve(s.listener)
	if err != nil {
		s.logger.Fatalf("Error while serving: %s", err)
	}
}

func (s *Service) Close() (err error) {
	s.listener.Close()
	err = s.server.Shutdown(s.ctx)
	return
}
