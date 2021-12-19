package main

import (
	"context"
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
	envPath                = "cache.env"
	balanceChannelCapacity = 10
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
	var err error
	var frequency int

	// Check for frequency of update arg
	args := os.Args

	if len(args) < 2 {
		log.Println("No update frequency provided - setting to default value of 5 seconds.")
	} else if frequency, err = strconv.Atoi(args[1]); err != nil {
		log.Println("Invalid update frequency provided - setting to default value of 5 seconds.")
	}
	if frequency == 0 {
		frequency = 5
	}

	// Get env variables and set a config
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Environmental variables failed to load: %s\n", err)
	}

	host := os.Getenv("HOST")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Unable to read environmental port variable: %s", err)
	}
	config := &Config{
		Host: host,
		Port: port,
	}

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize cache and DB connection
	bankBalancesChannel := make(chan *utils.BankBalance)
	accountBalancesChannel := make(chan *utils.Balance) // For the cache to send data to the db.

	bankRetreivalChannel := make(chan *utils.BankBalance)
	accountRetreivalChannel := make(chan *utils.Balance) // To retreive balances from db.

	// The dbstore must be initialized concurrently to correctly update the cache with retreived db balances.
	go func() {
		err = dbstore.New(
			ctx,
			bankBalancesChannel,
			accountBalancesChannel,
			bankRetreivalChannel,
			accountRetreivalChannel,
		)
		if err != nil {
			log.Fatalf("Could not init DB connection: %s", err)
		}
	}()

	err = memstore.New(
		ctx,
		bankBalancesChannel,
		accountBalancesChannel,
		bankRetreivalChannel,
		accountRetreivalChannel,
	)
	if err != nil {
		log.Fatalf("Could not init memstore: %s", err)
	}

	// Initialize service
	httpService, err := service.New(config, ctx)
	if err != nil {
		log.Fatalf("Failed to initialize service: %s", err)
	}
	log.Println("HTTP service started successfully.")

	// Display updated bank balances every 5 seconds
	go func() {
		log.Println("Transactions processed:")
		ticker := time.NewTicker(time.Duration(frequency) * time.Second)
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
		log.Println("Signal received, shutting down...")
		break
	case <-ctx.Done():
		log.Println("Main context cancelled, shutting down...")
		break

	}

	// First close memstore so it updates through db connection
	memstore.Close()
	if err != nil {
		log.Fatalf("Failed to close memstore: %s", err)
	}

	// Then close db connection
	dbstore.Close()
	if err != nil {
		log.Fatalf("Failed to close db connection: %s", err)
	}

	// Lastly, close service
	err = httpService.Close()
	if err != nil {
		log.Fatalf("Failed to close service: %s", err)
	}
}
