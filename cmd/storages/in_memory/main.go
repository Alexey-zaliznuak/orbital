// In-memory storage service — HTTP-хранилище сообщений в оперативной памяти.
package main

import (
	"context"
	"log"
	"time"

	inmemory "github.com/Alexey-zaliznuak/orbital/internal/storages/in_memory"
	"github.com/Alexey-zaliznuak/orbital/pkg/httputil"
)

func main() {
	cfg := inmemory.NewBuilder().FromEnv().Build()

	log.Printf("Starting in-memory storage server...")
	log.Printf("Storage ID: %s", cfg.ID)
	log.Printf("HTTP addr: %s", cfg.Address)

	store := inmemory.NewInMemoryStorage()
	if err := store.Initialize(context.Background(), cfg); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	server := inmemory.NewServer(store, inmemory.ServerConfig{
		Addr:         cfg.Address,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	log.Printf("HTTP server listening on %s", cfg.Address)
	httputil.Run(server, 10*time.Second)
	log.Printf("In-memory storage stopped")
}
