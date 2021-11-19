package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IsakJones/polka/api/cache"
	"github.com/IsakJones/polka/api/service"
	"github.com/IsakJones/polka/api/service/handlers"
)

const (
	updateFrequency = 5 * time.Second
	defaultPort     = ":8090"
)

func main() {

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service
	httpService := service.New(defaultPort, ctx)
	if err := httpService.Start(); err != nil {
		log.Fatalf("HTTP service failed to start: %s", err)
	}
	log.Println("HTTP service started successfully.")

	// Update state every 5 seconds
	go func() {
		fmt.Println("Transactions processed:")
		ticker := time.NewTicker(updateFrequency)
		for range ticker.C {
			handlers.PrintProcessedTransactions()
			cache.PrintDues()
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	// Block until a SIGTERM comes through
	select {
	case <-signalChannel:
		log.Println("Signal received, shutting down...")
		// TODO
	}
}
