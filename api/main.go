package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/sekerez/polka/api/client"
	"github.com/sekerez/polka/api/dbstore"
	"github.com/sekerez/polka/api/service"
	"github.com/sekerez/polka/api/service/handlers"
)

const (
	mainEnv             = "api.env"
	cacheEnv            = "cache.env"
	cacheConnTimeout    = 30 * time.Second
	cacheReqTimeout     = 10 * time.Second
	balanceInterval     = 1
	balanceChanCapacity = 20
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

/*
TODOS:
 * Play with indexes to see if it increases DB performance
 * Deploy to cloud nodes
 * Run over longer dataset to establish consistant QPS
 * Look at system metrics to see load
   * Is it just too much concurrency, or is db under too much load
*/
func main() {

	// Initialize logger
	logger := log.New(os.Stderr, "[main] ", log.LstdFlags|log.Lshortfile)

	// Parse frequency flag
	frequencyPtr := flag.Int("f", 5, "update frequency")
	flag.Parse()
	frequency := time.Duration(*frequencyPtr) * time.Second

	logger.Printf(frequency.String())

	// Get env variables and set a config
	if err := godotenv.Load(mainEnv); err != nil {
		logger.Fatalf("Environmental variables failed to load: %s\n", err)
	}

	host := os.Getenv("HOST")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		logger.Fatalf("Unable to read environmental port variable: %s", err)
	}
	config := &Config{
		Host: host,
		Port: port,
	}

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize client
	if err := godotenv.Load(cacheEnv); err != nil {
		logger.Fatalf("Cache environmental variables failed to load: %s\n", err)
	}
	err = client.New(
		os.Getenv("CACHEADDRESS"),
		cacheConnTimeout,
		cacheReqTimeout,
	)
	if err != nil {
		logger.Fatalf("Could not start client: %s", err)
	}

	// Initialize database connection
	err = dbstore.New(ctx)
	if err != nil {
		logger.Fatalf("Could not init DB connection: %s", err)
	}

	// Initialize service
	httpService, err := service.New(config, ctx)
	if err != nil {
		logger.Fatalf("Failed to initialize service: %s", err)
	}
	logger.Println("HTTP service started successfully.")

	// Print processed transactions regularly
	go func() {
		ticker := time.NewTicker(frequency)
		for range ticker.C {
			handlers.PrintProcessedTransactions()
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

	err = httpService.Close()
	if err != nil {
		logger.Fatalf("Failed to close service: %s", err)
	}
}
