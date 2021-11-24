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

	"github.com/IsakJones/polka/api/memstore"
	"github.com/IsakJones/polka/api/service"
	"github.com/IsakJones/polka/api/service/handlers"
	"github.com/IsakJones/polka/api/utils"
)

const (
	envPath = "api.env"
)

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
		log.Fatalf("Unable to read environmental port variable: %s", port)
	}
	config := &utils.Config{
		Host: host,
		Port: port,
	}

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service
	httpService := service.New(config, ctx)
	if err := httpService.Start(); err != nil {
		log.Fatalf("HTTP service failed to start: %s\n", err)
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
