// Coordinator service — управляет состоянием кластера.
package main

import (
	"context"
	"log"
	"time"

	"github.com/Alexey-zaliznuak/orbital/internal/coordinator"
	"github.com/Alexey-zaliznuak/orbital/internal/coordinator/config"
	coordinatorhttp "github.com/Alexey-zaliznuak/orbital/internal/coordinator/http"
	"github.com/Alexey-zaliznuak/orbital/internal/coordinator/storage/etcd"
	"github.com/Alexey-zaliznuak/orbital/pkg/httputil"

	_ "github.com/Alexey-zaliznuak/orbital/docs/swagger-coordinator"
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

	log.Printf("HTTP server listening on %s", coordinatorConfig.HTTPAddr)
	httputil.Run(server, 10*time.Second)
	log.Printf("Coordinator stopped")
}
