package inmemory

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
	"github.com/caarlos0/env/v11"
)

type InMemoryStorageConfig struct {
	storage.BaseStorageConfig

	UseDump  bool   `env:"USE_DUMP" envDefault:"true"`
	DumpFile string `env:"DUMP_FILE" envDefault:"./dump.json"`
}

type InMemoryStorageConfigBuilder struct {
	cfg *InMemoryStorageConfig
}

func NewBuilder() *InMemoryStorageConfigBuilder {
	return &InMemoryStorageConfigBuilder{
		cfg: &InMemoryStorageConfig{},
	}
}

func (b *InMemoryStorageConfigBuilder) WithID(id string) *InMemoryStorageConfigBuilder {
	b.cfg.ID = id
	return b
}

func (b *InMemoryStorageConfigBuilder) WithAddress(addr string) *InMemoryStorageConfigBuilder {
	b.cfg.Address = addr
	return b
}

func (b *InMemoryStorageConfigBuilder) WithClusterAddress(addr string) *InMemoryStorageConfigBuilder {
	b.cfg.ClusterAddress = addr
	return b
}

func (b *InMemoryStorageConfigBuilder) WithMinDelay(d time.Duration) *InMemoryStorageConfigBuilder {
	b.cfg.MinDelay = d
	return b
}

func (b *InMemoryStorageConfigBuilder) WithMaxDelay(d time.Duration) *InMemoryStorageConfigBuilder {
	b.cfg.MaxDelay = d
	return b
}

func (b *InMemoryStorageConfigBuilder) WithFindExpiredInterval(d time.Duration) *InMemoryStorageConfigBuilder {
	b.cfg.FindExpiredInterval = d
	return b
}

func (b *InMemoryStorageConfigBuilder) WithSendExpiredInterval(d time.Duration) *InMemoryStorageConfigBuilder {
	b.cfg.SendExpiredInterval = d
	return b
}

func (b *InMemoryStorageConfigBuilder) WithDumpFile(path string) *InMemoryStorageConfigBuilder {
	b.cfg.DumpFile = path
	return b
}

func (b *InMemoryStorageConfigBuilder) WithMaxOutputBatchSize(size int) *InMemoryStorageConfigBuilder {
	b.cfg.MaxOutputBatchSize = size
	return b
}

func (b *InMemoryStorageConfigBuilder) FromEnv() *InMemoryStorageConfigBuilder {
	env.Parse(b.cfg)
	return b
}

func (b *InMemoryStorageConfigBuilder) Build() *InMemoryStorageConfig {
	return b.cfg
}
