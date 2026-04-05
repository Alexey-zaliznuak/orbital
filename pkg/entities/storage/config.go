package storage

import "time"

type BaseStorageConfig struct {
	ID       string        `env:"IN_MEMORY_STORAGE_ID"        envDefault:"in-memory"`
	Address  string        `env:"IN_MEMORY_STORAGE_ADDRESS"   envDefault:""`
	MinDelay time.Duration `env:"IN_MEMORY_STORAGE_MIN_DELAY" envDefault:"0"`
	MaxDelay time.Duration `env:"IN_MEMORY_STORAGE_MAX_DELAY" envDefault:"0"`
}
