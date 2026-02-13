package gateway

type GatewayConfig struct {
	ClusterAddress string `json:"cluster_address"`

	HTTPAddr string `json:"http_addr"`
	GRPCAddr string `json:"grpc_addr"`

	LogLevel string `json:"log_level"`
}
