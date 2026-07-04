package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/app"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/config"
)

func main() {

	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    cfg.HTTPAddress(),
		Handler: application.Router,
	}

	go func() {
		log.Printf("Server listening on %s\n", cfg.HTTPAddress())

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped.")
}
