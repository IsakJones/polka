package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IsakJones/polka/api/memstore"
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
	err := httpService.Close()
	if err != nil {
		log.Fatalf("Failed to close service: %s", err)
	}
}