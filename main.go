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

	"frith/common"
)

func main() {
	common.SetupEnvironment()

	// Create context that we'll cancel on shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run GC
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
			again:
				err := common.GetDatabase().RunValueLogGC(0.7)
				if err == nil {
					goto again
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	defer ticker.Stop()
	defer common.GetDatabase().Close()

	// Relay

	relay := common.GetRelay()

	// Merge everything into a single handler and start the server

	mux := relay.Router()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `This is a nostr relay, please connect using wss://`)
	})

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", common.PORT),
		Handler: relay,
	}

	// Start server in goroutine
	go func() {
		fmt.Printf("running on :%s\n", common.PORT)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown the HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	}

	// Cancel context to stop background tasks
	cancel()
}
