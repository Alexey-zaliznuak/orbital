package coordinator

import "time"

type CoordinatorConfig struct {
	// HTTP
	HTTPAddr         string        `env:"HTTP_ADDR"          envDefault:":8080"`
	HTTPReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT"  envDefault:"15s"`
	HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`

	// etcd
	EtcdEndpoints   []string      `env:"ETCD_ENDPOINTS"    envDefault:"localhost:2379" envSeparator:","`
	EtcdDialTimeout time.Duration `env:"ETCD_DIAL_TIMEOUT" envDefault:"5s"`
	EtcdOpTimeout   time.Duration `env:"ETCD_OP_TIMEOUT"   envDefault:"5s"`

	// Логирование
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

type ClusterConfig struct {
	NatsAddress string `json:"nats_address" env:"NATS_URL" envDefault:"localhost:4222"`
}
