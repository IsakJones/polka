package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/sekerez/polka/processor/src/utils"
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
	ctx      context.Context
}

func (s *Service) Address() net.Addr {
	return s.listener.Addr()
}

// New returns an uninitialized http service.
func New(conf utils.Config, ctx context.Context) (*Service, error) {

	logger := log.New(os.Stderr, "[service] ", log.LstdFlags|log.Lshortfile)

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
	mux.HandleFunc(transPath, handleTransactions)
	mux.HandleFunc(helloPath, handleHello)

	// Set up server
	server := &http.Server{
		Handler: mux,
	}

	// Successfully initialize service
	s := &Service{
		ctx:      ctx,
		mux:      mux,
		conf:     conf,
		logger:   logger,
		server:   server,
		listener: listener,
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
