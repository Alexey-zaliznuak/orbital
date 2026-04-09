package storage

import "time"

type BaseStorageConfig struct {
	ID       string        `env:"STORAGE_ID"        envDefault:"in-memory"`

	ClusterAddress string `json:"cluster_address" env:"COORDINATOR_ADDR" envDefault:""`

	Address  string        `env:"STORAGE_ADDRESS"   envDefault:""`

	MinDelay time.Duration `env:"STORAGE_MIN_DELAY" envDefault:"0"`
	MaxDelay time.Duration `env:"STORAGE_MAX_DELAY" envDefault:"0"`

	FindExpiredInterval time.Duration `env:"FIND_EXPIRED_INTERVAL" envDefault:"10ms"`
	SendExpiredInterval time.Duration `env:"SEND_EXPIRED_INTERVAL" envDefault:"10ms"`

	MaxOutputBatchSize int `env:"MAX_OUTPUT_BATCH_SIZE" envDefault:"100"`
}
