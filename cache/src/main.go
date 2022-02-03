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

	"github.com/sekerez/polka/cache/src/dbstore"
	"github.com/sekerez/polka/cache/src/memstore"
	"github.com/sekerez/polka/cache/src/service"
	"github.com/sekerez/polka/cache/src/utils"
)

const (
	envPath = "env/cache.env"
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

	// Parse frequency flag
	frequencyPtr := flag.Int("f", 5, "update frequency")
	withAccPtr := flag.Bool("a", false, "print accounts with dues")
	flag.Parse()
	frequency := time.Duration(*frequencyPtr) * time.Second

	// Get env variables and set a config
	if err := godotenv.Load(envPath); err != nil {
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

	// Initialize cache and DB connection
	bankNumChan := make(chan uint16)
	bankBalancesChannel := make(chan *utils.BankBalance)
	accountBalancesChannel := make(chan *utils.Balance) // For the cache to send data to the db.

	bankRetreivalChannel := make(chan *utils.BankBalance)
	accountRetreivalChannel := make(chan *utils.Balance) // To retreive balances from db.

	// The dbstore must be initialized concurrently to correctly update the cache with retreived db balances.
	go func() {
		err = dbstore.New(
			ctx,
			bankNumChan,
			bankBalancesChannel,
			accountBalancesChannel,
			bankRetreivalChannel,
			accountRetreivalChannel,
		)
		if err != nil {
			logger.Fatalf("Could not init DB connection: %s", err)
		}
	}()

	// Initialize cache
	err = memstore.New(
		ctx,
		bankNumChan,
		bankBalancesChannel,
		accountBalancesChannel,
		bankRetreivalChannel,
		accountRetreivalChannel,
	)
	if err != nil {
		logger.Fatalf("Could not init memstore: %s", err)
	}

	// Initialize service
	s, err := service.New(u, ctx)
	if err != nil {
		logger.Fatalf("Failed to initialize service: %s", err)
	}
	logger.Printf("HTTP service initialized successfully.")

	// Listen for requests
	go func() {
		errChan := make(chan error)
		s.Serve(errChan)
		err = <-errChan
		if err != nil {
			logger.Printf("Error serving: %s", err)
		}
	}()

	// Display updated bank balances every 5 seconds
	go func() {
		logger.Println("Transactions processed:")
		ticker := time.NewTicker(frequency)
		for range ticker.C {
			memstore.PrintBalances(*withAccPtr)
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

	// First close memstore so it updates through db connection
	memstore.Close()
	if err != nil {
		logger.Fatalf("Failed to close memstore: %s", err)
	}
	logger.Printf("Shut down memory cache.")

	// Then close db connection
	dbstore.Close()
	if err != nil {
		logger.Fatalf("Failed to close db connection: %s", err)
	}
	logger.Printf("Closed DB connection.")

	// Lastly, close service
	err = s.Close()
	if err != nil {
		logger.Fatalf("Failed to shut down service: %s", err)
	}
	logger.Printf("Successfully shut down service.")
}
