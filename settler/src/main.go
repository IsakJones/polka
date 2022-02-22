package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/sekerez/polka/settler/src/client"
	"github.com/sekerez/polka/settler/src/service"
)

const (
	mainEnv         = "env/settler.env"
	cacheReqTimeout = 10 * time.Second
)

type Config struct {
	Host string
	Port int
}

func (c *Config) GetHost() string {
	return c.Host
}

func (c *Config) GetPort() int {
	return c.Port
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) GetListenPort() string {
	return fmt.Sprintf(":%d", c.Port)
}

func main() {

	// Initialize logger
	logger := log.New(os.Stderr, "[main] ", log.LstdFlags|log.Lshortfile)

	// Get env variables and set a config
	if err := godotenv.Load(mainEnv); err != nil {
		logger.Fatalf("Environmental variables failed to load: %s\n", err)
	}

	host := os.Getenv("HOST")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		logger.Fatalf("Unable to read environmental port variable: %s", err)
	}
	u, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		logger.Fatalf("Unable to parse url: %s", err)
	}

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize client
	err = client.New(os.Getenv("CACHEADDRESS"), cacheReqTimeout)
	if err != nil {
		logger.Fatalf("Could not start client: %s", err)
	}

	// Initialize service
	s, err := service.New(u, ctx)
	if err != nil {
		logger.Fatalf("Failed to initialize service: %s", err)
	}
	logger.Println("HTTP service initialized successfully.")

	// Listen for requests
	go func() {
		logger.Printf("Awaiting requests...")
		errChan := make(chan error)
		s.Serve(errChan)
		err = <-errChan
		if err != nil {
			logger.Printf("Error serving: %s", err)
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
	logger.Printf("Shut down api service.")
}
