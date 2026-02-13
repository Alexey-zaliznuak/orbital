package config

import (
	"github.com/Alexey-zaliznuak/orbital/pkg/config"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
)

// ClusterConfigBuilder для построения конфигурации.
type ClusterConfigBuilder struct {
	cfg *coordinator.ClusterConfig
}

// NewBuilder создаёт новый builder с дефолтными значениями.
func NewClusterConfigBuilder() *ClusterConfigBuilder {
	builder := &ClusterConfigBuilder{
		cfg: &coordinator.ClusterConfig{
			NatsAddress: "localhost:4222",
		},
	}
	return builder
}

// WithHTTPAddr устанавливает адрес HTTP сервера.
func (b *ClusterConfigBuilder) WithNatsAddress(addr string) *ClusterConfigBuilder {
	b.cfg.NatsAddress = addr
	return b
}

// FromEnv загружает конфигурацию из переменных окружения.
func (b *ClusterConfigBuilder) FromEnv() *ClusterConfigBuilder {
	b.cfg.NatsAddress = config.GetEnv("NATS_URL", config.GetEnv("NATS_URI", b.cfg.NatsAddress))

	return b
}

// Build возвращает готовую конфигурацию.
func (b *ClusterConfigBuilder) Build() *coordinator.ClusterConfig {
	return b.cfg
}
