package config

import (
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
	"github.com/caarlos0/env/v11"
)

// ClusterConfigBuilder для построения конфигурации.
type ClusterConfigBuilder struct {
	cfg *coordinator.ClusterConfig
}

// NewClusterConfigBuilder создаёт новый builder с дефолтными значениями.
func NewClusterConfigBuilder() *ClusterConfigBuilder {
	return &ClusterConfigBuilder{
		cfg: &coordinator.ClusterConfig{},
	}
}

// WithNatsAddress устанавливает адрес NATS сервера.
func (b *ClusterConfigBuilder) WithNatsAddress(addr string) *ClusterConfigBuilder {
	b.cfg.NatsAddress = addr
	return b
}

// FromEnv загружает конфигурацию из переменных окружения.
func (b *ClusterConfigBuilder) FromEnv() *ClusterConfigBuilder {
	env.Parse(b.cfg)
	return b
}

// Build возвращает готовую конфигурацию.
func (b *ClusterConfigBuilder) Build() *coordinator.ClusterConfig {
	return b.cfg
}
