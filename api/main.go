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

	"github.com/sekerez/polka/api/dbstore"
	"github.com/sekerez/polka/api/memstore"
	"github.com/sekerez/polka/api/service"
	"github.com/sekerez/polka/api/service/handlers"
)

const (
	envPath = "api.env"
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

	// Initialize database connection
	err = dbstore.New(ctx)
	if err != nil {
		log.Fatalf("Could not init DB connection: %s", err)
	}

	// Initialize service
	httpService, err := service.New(config, ctx)
	if err != nil {
		log.Fatalf("Failed to initialize service: %s", err)
	}
	log.Println("HTTP service started successfully.")

	// Update state every 5 seconds
	go func() {
		fmt.Println("Transactions processed:")
		ticker := time.NewTicker(time.Duration(frequency) * time.Second)
		for range ticker.C {
			handlers.PrintProcessedTransactions()
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

	err = httpService.Close()
	if err != nil {
		log.Fatalf("Failed to close service: %s", err)
	}
}
