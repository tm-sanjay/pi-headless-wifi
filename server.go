package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var server *http.Server

func initServer() {
	// Create an HTTP server
	server = &http.Server{
		Addr:    ":8080",
		Handler: nil, // Use the default router
	}

	fmt.Println("Starting server...")
	fmt.Println(ip + ":" + port)

	// Handle the routes
	http.HandleFunc("/", home)
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/wifilist", getWifiList)

	// Start the HTTP server in a separate goroutine
	go func() {
		log.Println("Http Server started (init)")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panic("Error starting the server:", err)
		}
	}()

	// Create a channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signal or timeout
	select {
	case <-stop:
		log.Println("Termination signal received.")
	case <-time.After(5 * time.Second):
		log.Println("Timeout reached.")
	}

	// Shutdown the server gracefully
	shutdownServer(server)
}

func startHttpServer() {
	// Start the HTTP server
	go func() {
		log.Println("Http Server started.")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panic("Error starting the server:", err)
		}
	}()
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Panic("Error shutting down the server:", err)
	}

	log.Println("Server stopped.")
}
