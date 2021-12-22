package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"

	"github.com/sekerez/polka/lb/utils"
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
	pool     *apiPool
	ctx      context.Context
}

type apiPool struct {
	servers []*apiServer
	current uint64
}

type apiServer struct {
	sync.RWMutex
	Url          *url.URL
	Alive        bool
	ReverseProxy *httputil.ReverseProxy
}

func (as *apiServer) revive() {
	as.Lock()
	defer as.Unlock()
	as.Alive = true
}

func (as *apiServer) kill() {
	as.Lock()
	defer as.Unlock()
	as.Alive = false
}

func (as *apiServer) isAlive() bool {
	as.Lock()
	defer as.Unlock()
	return as.Alive
}

// New returns an uninitialized http service.
func New(conf utils.Config, apiUrls []*url.URL, ctx context.Context) (*Service, error) {

	logger := log.New(os.Stderr, "[service] ", log.LstdFlags|log.Lshortfile)

	// Set up server
	server := &http.Server{
		Handler: http.HandlerFunc(handle),
		Addr:    conf.GetListenPort(),
	}

	// Configure TCP connection
	tcpAddr, err := net.ResolveTCPAddr("tcp4", conf.GetListenPort())
	if err != nil {
		return nil, err
	}

	// Set up listener
	listener, err := net.Listen("tcp", tcpAddr.String())
	if err != nil {
		return nil, err
	}

	// Set up serverpool
	pool := &apiPool{
		servers: make([]*apiServer, len(apiUrls)),
		current: 0,
	}

	// Set up api servers in server pool
	for i, u := range apiUrls {
		proxy := httputil.NewSingleHostReverseProxy(u)
		pool.servers[i] = &apiServer{
			ReverseProxy: proxy,
			Alive:        true,
			Url:          u,
		}
	}

	// Successfully initialize service
	s := &Service{
		ctx:      ctx,
		conf:     conf,
		logger:   logger,
		server:   server,
		listener: listener,
	}

	// Start listening for requests
	return s, nil
}

func handle(w http.ResponseWriter, r *http.Request) {

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
