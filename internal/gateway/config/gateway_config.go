package config

import (
	"github.com/Alexey-zaliznuak/orbital/pkg/config"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
)

type GatewayConfigBuilder struct {
	cfg *gateway.GatewayConfig
}

// NewBuilder создаёт новый builder с дефолтными значениями.
func NewGatewayConfigBuilder() *GatewayConfigBuilder {
	builder := &GatewayConfigBuilder{
		cfg: &gateway.GatewayConfig{
			ClusterAddress: "",
			HTTPAddr:       ":8080",
			GRPCAddr:       ":9090",
			LogLevel:       "info",
		},
	}
	return builder
}

// WithHTTPAddr устанавливает адрес HTTP сервера.
func (b *GatewayConfigBuilder) WithClusterAddress(addr string) *GatewayConfigBuilder {
	b.cfg.ClusterAddress = addr
	return b
}

// FromEnv загружает конфигурацию из переменных окружения.
func (b *GatewayConfigBuilder) FromEnv() *GatewayConfigBuilder {
	b.cfg.ClusterAddress = config.GetEnv("COORDINATOR_ADDR", b.cfg.ClusterAddress)

	b.cfg.HTTPAddr = config.GetEnv("HTTP_ADDR", b.cfg.HTTPAddr)
	b.cfg.GRPCAddr = config.GetEnv("GRPC_ADDR", b.cfg.GRPCAddr)

	b.cfg.LogLevel = config.GetEnv("LOG_LEVEL", b.cfg.LogLevel)

	return b
}

// Build возвращает готовую конфигурацию.
func (b *GatewayConfigBuilder) Build() *gateway.GatewayConfig {
	return b.cfg
}
