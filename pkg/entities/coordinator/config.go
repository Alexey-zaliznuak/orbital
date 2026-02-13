package coordinator

import "time"

type CoordinatorConfig struct {
	// HTTP
	HTTPAddr         string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration

	// etcd
	EtcdEndpoints   []string
	EtcdDialTimeout time.Duration
	EtcdOpTimeout   time.Duration

	// Логирование
	LogLevel string
}

type ClusterConfig struct {
	NatsAddress string `json:"nats_address"`
}
