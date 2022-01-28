package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"
)

const (
	transPath          = "/transaction"
	checkTimeout       = 2 * time.Second
	Attempts     uint8 = iota
	Retries
)

var pool *apiPool

// Service manages the main application functions.
type Service struct {
	logger      *log.Logger
	listener    net.Listener
	server      *http.Server
	ctx         context.Context
	quit        chan struct{}
	checkIsDone chan struct{}
}

// New returns an uninitialized http service.
func New(lbUrl *url.URL, apiUrls []*url.URL, ctx context.Context) (*Service, error) {

	logger := log.New(os.Stderr, "[service] ", log.LstdFlags|log.Lshortfile)

	logger.Printf("urls. size: %v", len(apiUrls))

	// Set up server
	server := &http.Server{
		Handler: http.HandlerFunc(handle),
		Addr:    fmt.Sprintf(":%s", lbUrl.Port()),
	}

	// Configure TCP connection
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%s", lbUrl.Port()))

	if err != nil {
		return nil, err
	}

	// Set up listener
	listener, err := net.Listen("tcp", tcpAddr.String())
	if err != nil {
		return nil, err
	}

	// Set up api servers after initializing apiPool
	pool = &apiPool{}
	for _, u := range apiUrls {
		logger.Printf("url: host: %v, port: %v", u.Host, u.Port())
		pool.add(u)
	}

	// Where is the channel created?
	// Successfully initialize service
	s := &Service{
		ctx:         ctx,
		logger:      logger,
		server:      server,
		listener:    listener,
		quit:        make(chan struct{}),
		checkIsDone: make(chan struct{}),
	}

	// Start listening for requests
	return s, nil
}

func handle(w http.ResponseWriter, r *http.Request) {
	attempts := getAttempts(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	api, err := pool.nextApi()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		log.Printf("Cannot provide service: %s", err.Error())
		return
	}

	atomic.AddUint64(&pool.counter, 1)
	api.reverseProxy.ServeHTTP(w, r)
}

// Start sets up a server and listener for incoming requests.
func (s *Service) Serve(errChan chan<- error) {
	// Periodically check if apis are alive/dead
	go s.healthCheck()
	errChan <- s.server.Serve(s.listener)
}

func (s *Service) PrintRequestNumber() {
	s.logger.Printf("%d transactions forwarded", pool.counter)
}

func (s *Service) Close() (err error) {
	// Wait for healthcheck to end
	s.logger.Printf("top of s.Close()")
	s.quit <- struct{}{}
	s.logger.Printf("waiting is checkisdone")
	<-s.checkIsDone
	// Close listener and server
	s.listener.Close()
	err = s.server.Shutdown(s.ctx)
	return
}
