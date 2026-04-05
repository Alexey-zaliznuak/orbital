// Gateway service — принимает сообщения и распределяет их по storage нодам.
package main

import (
	"context"
	"log"
	"time"

	"github.com/Alexey-zaliznuak/orbital/internal/gateway"
	"github.com/Alexey-zaliznuak/orbital/internal/gateway/config"
	gatewayhttp "github.com/Alexey-zaliznuak/orbital/internal/gateway/http"
	"github.com/Alexey-zaliznuak/orbital/pkg/httputil"
	"github.com/Alexey-zaliznuak/orbital/pkg/logger"

	_ "github.com/Alexey-zaliznuak/orbital/docs/swagger-gateway" // Swagger docs
)

func main() {
	// Загрузка конфигурации
	cfg := config.NewGatewayConfigBuilder().FromEnv().Build()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	log.Printf("Starting gateway server...")
	log.Printf("HTTP addr: %s", cfg.HTTPAddr)
	log.Printf("Cluster address: %s", cfg.ClusterAddress)

	// Создание gateway
	ctx := context.Background()
	gw, err := gateway.NewBaseGateway(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}

	gw.Start(ctx)

	// Создание HTTP сервера
	server := gatewayhttp.NewServer(gw, gatewayhttp.Config{
		Addr:         cfg.HTTPAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
	httputil.Run(server, 10*time.Second)
	log.Printf("Gateway stopped")
}
