package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/sekerez/polka/balancer/service"
)

const (
	mainEnv        = "lb.env"
	apiConnTimeout = 30 * time.Second
	apiReqTimeout  = 10 * time.Second
)

type Config struct {
	Host string
	Port int
	Url  *url.URL
}

func (c *Config) GetAddress() *url.URL {
	return c.Url
}

func (c *Config) GetListenPort() string {
	return fmt.Sprintf(":%d", c.Port)
}

func main() {

	// Initialize logger
	logger := log.New(os.Stderr, "[main] ", log.LstdFlags|log.Lshortfile)

	// Parse frequency flag
	frequencyPtr := flag.Int("f", 2, "update frequency")
	flag.Parse()
	frequency := time.Duration(*frequencyPtr) * time.Second

	// Get env variables
	if err := godotenv.Load(mainEnv); err != nil {
		logger.Fatalf("Environmental variables failed to load: %s\n", err)
	}

	// Set config
	host := os.Getenv("HOST")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		logger.Fatalf("Unable to read environmental port variable: %s", err)
	}
	u, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		logger.Fatalf("Unable to parse url: %s", err)
	}

	// Get API addresses
	apiNum, err := strconv.Atoi(os.Getenv("APINUM"))
	if err != nil {
		logger.Fatalf("Unable to read environmental API address variables: %s", err)
	}
	apiUrls := make([]*url.URL, apiNum)
	for i := 0; i < apiNum; i++ {
		apiUrl, err := url.Parse(os.Getenv(fmt.Sprintf("APIADDRESS%d", i)))
		if err != nil {
			logger.Printf("Failed to parse %d-th api server: %s", i, err)
		}
		apiUrls[i] = apiUrl
	}

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize service
	s, err := service.New(u, apiUrls, ctx)
	if err != nil {
		logger.Fatalf("Failed to initialize service: %s", err)
	}
	logger.Println("HTTP service initialized successfully.")

	// Listen for requests
	go func() {
		errChan := make(chan error)
		s.Serve(errChan)
		err = <-errChan
		if err != nil {
			logger.Printf("Error serving: %s", err)
		}
	}()

	// Print received requests
	go func() {
		ticker := time.NewTicker(frequency)
		for range ticker.C {
			s.PrintRequestNumber()
		}
	}()

	// Pipe incoming OS signals to channel
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	// Block until a SIGTERM comes through or the context shuts down
	select {
	case <-signalChannel:
		logger.Println("Signal received, shutting down...")
		break
	case <-ctx.Done():
		logger.Println("Main context cancelled, shutting down...")
		break
	}

	err = s.Close()
	if err != nil {
		logger.Fatalf("Failed to close service: %s", err)
	}
	logger.Printf("Successfully shut down load balancer.")
}
