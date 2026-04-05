package config

import (
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
	"github.com/caarlos0/env/v11"
)

type GatewayConfigBuilder struct {
	cfg *gateway.GatewayConfig
}

// NewGatewayConfigBuilder создаёт новый builder с дефолтными значениями.
func NewGatewayConfigBuilder() *GatewayConfigBuilder {
	return &GatewayConfigBuilder{
		cfg: &gateway.GatewayConfig{},
	}
}

// WithClusterAddress устанавливает адрес coordinator-а.
func (b *GatewayConfigBuilder) WithClusterAddress(addr string) *GatewayConfigBuilder {
	b.cfg.ClusterAddress = addr
	return b
}

// FromEnv загружает конфигурацию из переменных окружения.
func (b *GatewayConfigBuilder) FromEnv() *GatewayConfigBuilder {
	env.Parse(b.cfg)
	return b
}

// Build возвращает готовую конфигурацию.
func (b *GatewayConfigBuilder) Build() *gateway.GatewayConfig {
	return b.cfg
}
