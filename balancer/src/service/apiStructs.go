package service

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

// NB: url and reverse proxy should not be references if they are indeed static
type apiNode struct {
	alive        uint32
	url          url.URL
	reverseProxy httputil.ReverseProxy
}

func (an *apiNode) revive() {
	atomic.StoreUint32(&an.alive, 1)
}

func (an *apiNode) kill() {
	atomic.StoreUint32(&an.alive, 0)
}

func (an *apiNode) isAlive() bool {
	// Is this threadsafe?
	return an.alive == 1
}

type apiPool struct {
	apiNodes []*apiNode
	current  uint64 // The current index
	counter  uint64 // The number of transactions forwarded
}

// add adds an api node to the apiPool.
func (pool *apiPool) add(apiUrl *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(apiUrl)
	// proxy.ErrorHandler = proxyErrorFunc(proxy, apiUrl)
	log.Printf("%+v", apiUrl)

	node := &apiNode{
		url:          *apiUrl,
		alive:        1,
		reverseProxy: *proxy,
	}
	pool.apiNodes = append(pool.apiNodes, node)
}

// nextApi looks for the next alive backend in the list.
func (pool *apiPool) nextApi() (*apiNode, error) {

	// Increment current
	nextIndexUint64 := atomic.AddUint64(&pool.current, uint64(1))
	// Get next index
	nextIndex := int(nextIndexUint64) % len(pool.apiNodes)
	limit := len(pool.apiNodes) + nextIndex
	for i := nextIndex; i < limit; i++ {
		ind := i % len(pool.apiNodes)
		// Skip if the service is dead
		if !pool.apiNodes[ind].isAlive() {
			continue
		}
		// Update the atomic again if you skipped over nextIndex
		if ind != nextIndex {
			atomic.StoreUint64(&pool.current, uint64(ind))
		}
		return pool.apiNodes[ind], nil
	}

	err := errors.New("all api servers are dead")
	return nil, err
}

// killNode sets the given node as alive.
func (pool *apiPool) reviveNode(apiUrl *url.URL) error {
	for _, api := range pool.apiNodes {
		if &api.url == apiUrl {
			api.revive()
			return nil
		}
	}
	return errors.New("url not in api pool")
}

// killNode sets the given node as dead.
func (pool *apiPool) killNode(apiUrl *url.URL) error {
	for _, api := range pool.apiNodes {
		if api.url == *apiUrl {
			api.kill()
			return nil
		}
	}
	return errors.New("url not in api pool")
}

// proxyErrorFunc returns an error handling function that kills inactive nodes and forwards requests to alive nodes.
func proxyErrorFunc(proxy *httputil.ReverseProxy, apiUrl *url.URL) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("[%s] %s", apiUrl.Host, e.Error())

		// If there were fewer than 3 retries, try again
		if retries := getRetries(r); retries < 3 {
			// Give it 100 milliseconds
			<-time.After(100 * time.Millisecond)
			// Create new context
			ctx := context.WithValue(r.Context(), Retries, retries+1)
			proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// After three tries, kill the node
		err := pool.killNode(apiUrl)
		if err != nil {
			log.Printf("Could not find the following url while trying to kill: %s", apiUrl.String())
		}

		// Mark attempts
		attempts := getAttempts(r)
		log.Printf("Attempting retry #%d for request %s - %s\n", attempts, r.RemoteAddr, r.URL.Path)
		ctx := context.WithValue(r.Context(), Attempts, attempts+1)

		// Handle request again
		handle(w, r.WithContext(ctx))
	}

}

// getRetries returns the number of retries to the same api node for a request.
func getRetries(r *http.Request) uint8 {
	if retries, ok := r.Context().Value(Retries).(uint8); ok {
		return retries
	}
	return 0
}

// getAttempts return the number of attempts to connect to a api node for a request.
func getAttempts(r *http.Request) uint8 {
	if attempts, ok := r.Context().Value(Attempts).(uint8); ok {
		return attempts
	}
	return 1
}

// nodeResponds checks whether the api node is alive by establishing a TCP connection.
func nodeResponds(apiUrl *url.URL) bool {
	conn, err := net.DialTimeout("tcp", apiUrl.Host, checkTimeout/2)
	if err != nil {
		log.Printf("Url %s:%s unreachable: %s", apiUrl.Host, apiUrl.Port(), err.Error())
		return false
	}
	err = conn.Close()
	if err != nil {
		log.Printf("Error closing connection with healthy node: %s", err.Error())
		return false
	}
	return true
}

// healthCheck updates life status for each api node every checkTimeout seconds.
func (s *Service) healthCheck() {
	ticker := time.NewTicker(checkTimeout)
	for {
		select {
		// Quit gracefully
		case <-s.quit:
			// TODO Kill the nodes?
			log.Printf("got quit message")
			s.checkIsDone <- struct{}{}
			return
		case <-ticker.C:
			// Reset current to decrease modulo time
			atomic.StoreUint64(&pool.current, 0)
			s.logger.Printf("Starting health check...")
			for i, api := range pool.apiNodes {
				if nodeResponds(&api.url) != api.isAlive() {
					if api.isAlive() {
						api.kill()
					} else {
						api.revive()
					}
				}
				if api.isAlive() {
					s.logger.Printf("Api %d is alive.", i)
				} else {
					s.logger.Printf("Api %d is dead.", i)
				}
			}
			s.logger.Printf("Ended health check.")
		}
	}
}
