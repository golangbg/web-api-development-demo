package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/golangbg/web-api-development-demo/pkg/server"
)

func main() {
	// Get the address for our server from the environment
	addr := os.Getenv("BLOG_ADDR")
	if addr == "" {
		log.Printf("BLOG_ADDR is empty")
		os.Exit(1)
	}

	// Create a server instance
	srv, err := server.New(addr)
	if err != nil {
		// Something went wrong
		log.Printf("couldn't create server: %v", err)
		os.Exit(1)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Shutdown channel
	shutdown := make(chan struct{}, 1)

	// Launch the server via a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// Something went wrong. Print the error and return the shutdown signal.
			log.Printf("error: %v", err)
			shutdown <- struct{}{}
		}
	}()

	log.Print("ready to listen and serve")

	// Wait for anything to happen on the interrupt or shutdown channel
	select {
	case killSignal := <-interrupt:
		switch killSignal {
		// We got signalled on the interrupt channel
		case os.Interrupt:
			log.Print("got SIGINT...")
		case syscall.SIGTERM:
			log.Print("got SIGTERM...")
		}
	case <-shutdown:
		// We are forced to shutdown
		log.Printf("got a shutdown signal...")
	}

	log.Print("shutting down...")
	srv.Close()
	log.Print("done shutting down")
}
