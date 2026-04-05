package httputil

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server — интерфейс HTTP-сервера, поддерживающего graceful shutdown.
type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}

// Run запускает сервер и блокируется до получения сигнала остановки (SIGINT/SIGTERM).
// После сигнала выполняет graceful shutdown с указанным таймаутом.
func Run(server Server, shutdownTimeout time.Duration) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-done
	log.Printf("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
