package inmemory

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
	"github.com/caarlos0/env/v11"
)

// InMemoryStorageConfigBuilder строит конфигурацию in-memory хранилища.
type InMemoryStorageConfigBuilder struct {
	cfg *storage.BaseStorageConfig
}

// NewBuilder создаёт новый builder с дефолтными значениями.
func NewBuilder() *InMemoryStorageConfigBuilder {
	return &InMemoryStorageConfigBuilder{
		cfg: &storage.BaseStorageConfig{},
	}
}

// WithID устанавливает идентификатор хранилища.
func (b *InMemoryStorageConfigBuilder) WithID(id string) *InMemoryStorageConfigBuilder {
	b.cfg.ID = id
	return b
}

// WithAddress устанавливает адрес хранилища.
func (b *InMemoryStorageConfigBuilder) WithAddress(addr string) *InMemoryStorageConfigBuilder {
	b.cfg.Address = addr
	return b
}

// WithMinDelay устанавливает минимальную задержку сообщений.
func (b *InMemoryStorageConfigBuilder) WithMinDelay(d time.Duration) *InMemoryStorageConfigBuilder {
	b.cfg.MinDelay = d
	return b
}

// WithMaxDelay устанавливает максимальную задержку сообщений (0 = без ограничения).
func (b *InMemoryStorageConfigBuilder) WithMaxDelay(d time.Duration) *InMemoryStorageConfigBuilder {
	b.cfg.MaxDelay = d
	return b
}

// FromEnv загружает конфигурацию из переменных окружения.
func (b *InMemoryStorageConfigBuilder) FromEnv() *InMemoryStorageConfigBuilder {
	env.Parse(b.cfg)
	return b
}

// Build возвращает готовую конфигурацию.
func (b *InMemoryStorageConfigBuilder) Build() *storage.BaseStorageConfig {
	return b.cfg
}
