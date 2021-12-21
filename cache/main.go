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

	"github.com/sekerez/polka/cache/dbstore"
	"github.com/sekerez/polka/cache/memstore"
	"github.com/sekerez/polka/cache/service"
	"github.com/sekerez/polka/cache/utils"
)

const (
	envPath = "cache.env"
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
	flag.Parse()
	frequency := time.Duration(*frequencyPtr) * time.Second

	logger.Printf(frequency.String())

	// Get env variables and set a config
	if err := godotenv.Load(envPath); err != nil {
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
	httpService, err := service.New(config, ctx)
	if err != nil {
		logger.Fatalf("Failed to initialize service: %s", err)
	}
	logger.Println("HTTP service started successfully.")

	// Display updated bank balances every 5 seconds
	go func() {
		logger.Println("Transactions processed:")
		ticker := time.NewTicker(frequency)
		for range ticker.C {
			memstore.PrintDues()
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

	// Then close db connection
	dbstore.Close()
	if err != nil {
		logger.Fatalf("Failed to close db connection: %s", err)
	}

	// Lastly, close service
	err = httpService.Close()
	if err != nil {
		logger.Fatalf("Failed to close service: %s", err)
	}
}
