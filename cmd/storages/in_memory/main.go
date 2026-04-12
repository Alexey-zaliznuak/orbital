// In-memory storage service — HTTP-хранилище сообщений в оперативной памяти.
package main

import (
	"context"
	"log"
	"time"

	_ "github.com/Alexey-zaliznuak/orbital/docs/swagger-in-memory" // Swagger docs
	inmemory "github.com/Alexey-zaliznuak/orbital/internal/storages/in_memory"
	"github.com/Alexey-zaliznuak/orbital/pkg/httputil"
)

func main() {
	cfg := inmemory.NewBuilder().FromEnv().Build()
	ctx := context.Background()

	log.Printf("Starting in-memory storage server...")
	log.Printf("Storage ID: %s", cfg.ID)
	log.Printf("HTTP port: 8080")

	store := inmemory.NewInMemoryStorage()

	if err := store.Initialize(ctx, cfg); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	server := inmemory.NewServer(store, inmemory.ServerConfig{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	log.Printf("HTTP server listening on :8080")
	httputil.Run(server, 10*time.Second)
	log.Printf("In-memory storage stopped")
}
