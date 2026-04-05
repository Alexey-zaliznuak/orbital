package config

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
	"github.com/caarlos0/env/v11"
)

// CoordinatorConfigBuilder для построения конфигурации.
type CoordinatorConfigBuilder struct {
	cfg *coordinator.CoordinatorConfig
}

// NewCoordinatorConfigBuilder создаёт новый builder с дефолтными значениями.
func NewCoordinatorConfigBuilder() *CoordinatorConfigBuilder {
	return &CoordinatorConfigBuilder{
		cfg: &coordinator.CoordinatorConfig{},
	}
}

// WithHTTPAddr устанавливает адрес HTTP сервера.
func (b *CoordinatorConfigBuilder) WithHTTPAddr(addr string) *CoordinatorConfigBuilder {
	b.cfg.HTTPAddr = addr
	return b
}

// WithHTTPTimeouts устанавливает таймауты HTTP сервера.
func (b *CoordinatorConfigBuilder) WithHTTPTimeouts(read, write time.Duration) *CoordinatorConfigBuilder {
	b.cfg.HTTPReadTimeout = read
	b.cfg.HTTPWriteTimeout = write
	return b
}

// WithEtcdEndpoints устанавливает адреса etcd.
func (b *CoordinatorConfigBuilder) WithEtcdEndpoints(endpoints []string) *CoordinatorConfigBuilder {
	b.cfg.EtcdEndpoints = endpoints
	return b
}

// WithEtcdTimeouts устанавливает таймауты etcd.
func (b *CoordinatorConfigBuilder) WithEtcdTimeouts(dial, op time.Duration) *CoordinatorConfigBuilder {
	b.cfg.EtcdDialTimeout = dial
	b.cfg.EtcdOpTimeout = op
	return b
}

// WithLogLevel устанавливает уровень логирования.
func (b *CoordinatorConfigBuilder) WithLogLevel(level string) *CoordinatorConfigBuilder {
	b.cfg.LogLevel = level
	return b
}

// FromEnv загружает конфигурацию из переменных окружения.
func (b *CoordinatorConfigBuilder) FromEnv() *CoordinatorConfigBuilder {
	env.Parse(b.cfg)
	return b
}

// Build возвращает готовую конфигурацию.
func (b *CoordinatorConfigBuilder) Build() *coordinator.CoordinatorConfig {
	return b.cfg
}
