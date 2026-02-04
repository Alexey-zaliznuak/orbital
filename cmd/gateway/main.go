// Coordinator service — управляет состоянием кластера.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alexey-zaliznuak/orbital/internal/coordinator"
	"github.com/Alexey-zaliznuak/orbital/internal/coordinator/config"
	coordinatorhttp "github.com/Alexey-zaliznuak/orbital/internal/coordinator/http"
	"github.com/Alexey-zaliznuak/orbital/internal/coordinator/storage/etcd"

	_ "github.com/Alexey-zaliznuak/orbital/docs/swagger" // Swagger docs
)

func main() {
	// Загрузка конфигурации
	coordinatorConfig := config.NewCoordinatorConfigBuilder().FromEnv().Build()
	clusterConfig := config.NewClusterConfigBuilder().FromEnv().Build()

	log.Printf("Starting coordinator server...")
	log.Printf("HTTP addr: %s", coordinatorConfig.HTTPAddr)
	log.Printf("etcd endpoints: %v", coordinatorConfig.EtcdEndpoints)
	log.Printf("nats address: %s", clusterConfig.NatsAddress)

	// Подключение к etcd
	storage, err := etcd.New(etcd.Config{
		Endpoints:   coordinatorConfig.EtcdEndpoints,
		DialTimeout: coordinatorConfig.EtcdDialTimeout,
		OpTimeout:   coordinatorConfig.EtcdOpTimeout,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
	defer storage.Close()

	log.Printf("Connected to etcd")

	// Создание координатора
	ctx := context.Background()
	coord, err := coordinator.NewBaseCoordinator(ctx, storage, coordinatorConfig, clusterConfig)
	if err != nil {
		log.Fatalf("Failed to create coordinator: %v", err)
	}

	// Создание HTTP сервера
	server := coordinatorhttp.NewServer(coord, coordinatorhttp.Config{
		Addr:         coordinatorConfig.HTTPAddr,
		ReadTimeout:  coordinatorConfig.HTTPReadTimeout,
		WriteTimeout: coordinatorConfig.HTTPWriteTimeout,
	})

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("HTTP server listening on %s", coordinatorConfig.HTTPAddr)
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Ожидание сигнала остановки
	<-done
	log.Printf("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Printf("Coordinator stopped")
}
