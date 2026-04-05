package gateway

type GatewayConfig struct {
	ClusterAddress string `json:"cluster_address" env:"COORDINATOR_ADDR" envDefault:""`

	HTTPAddr string `json:"http_addr" env:"HTTP_ADDR" envDefault:":8080"`
	GRPCAddr string `json:"grpc_addr" env:"GRPC_ADDR" envDefault:":9090"`

	LogLevel string `json:"log_level" env:"LOG_LEVEL" envDefault:"info"`
}
