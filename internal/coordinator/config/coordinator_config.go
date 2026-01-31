package config

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/config"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
)

// CoordinatorConfigBuilder для построения конфигурации.
type CoordinatorConfigBuilder struct {
	cfg *coordinator.CoordinatorConfig
}

// NewBuilder создаёт новый builder с дефолтными значениями.
func NewCoordinatorConfigBuilder() *CoordinatorConfigBuilder {
	builder := &CoordinatorConfigBuilder{
		cfg: &coordinator.CoordinatorConfig{
			HTTPAddr:         ":8080",
			HTTPReadTimeout:  15 * time.Second,
			HTTPWriteTimeout: 15 * time.Second,

			EtcdEndpoints:   []string{"localhost:2379"},
			EtcdDialTimeout: 5 * time.Second,
			EtcdOpTimeout:   5 * time.Second,

			LogLevel: "info",
		},
	}
	return builder
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
	b.cfg.HTTPAddr = config.GetEnv("HTTP_ADDR", b.cfg.HTTPAddr)
	b.cfg.HTTPReadTimeout = config.GetEnvDuration("HTTP_READ_TIMEOUT", b.cfg.HTTPReadTimeout)
	b.cfg.HTTPWriteTimeout = config.GetEnvDuration("HTTP_WRITE_TIMEOUT", b.cfg.HTTPWriteTimeout)

	b.cfg.EtcdEndpoints = config.GetEnvSlice("ETCD_ENDPOINTS", ",", b.cfg.EtcdEndpoints)
	b.cfg.EtcdDialTimeout = config.GetEnvDuration("ETCD_DIAL_TIMEOUT", b.cfg.EtcdDialTimeout)
	b.cfg.EtcdOpTimeout = config.GetEnvDuration("ETCD_OP_TIMEOUT", b.cfg.EtcdOpTimeout)

	b.cfg.LogLevel = config.GetEnv("LOG_LEVEL", b.cfg.LogLevel)

	return b
}

// Build возвращает готовую конфигурацию.
func (b *CoordinatorConfigBuilder) Build() *coordinator.CoordinatorConfig {
	return b.cfg
}
