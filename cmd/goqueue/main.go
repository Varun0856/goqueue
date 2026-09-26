package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Varun0856/goqueue/internal/api"
	"github.com/Varun0856/goqueue/internal/store"
	"github.com/Varun0856/goqueue/internal/worker"
)

func main() {
	s := store.NewStore()

	pool := worker.NewPool(4, 100, s)
	pool.Start()

	handler := api.NewHandler(pool, s)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", handler.CreateJob)
	mux.HandleFunc("GET /jobs/{id}", handler.GetJob)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		log.Println("Goqueue server listening on :8080")
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	pool.Stop()
	log.Println("Goqueue stopped")

}
