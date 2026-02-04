package gateway

type GatewayConfig struct {
	ClusterAddress string

	HTTPAddr string
	GRPCAddr string

	LogLevel string
}
