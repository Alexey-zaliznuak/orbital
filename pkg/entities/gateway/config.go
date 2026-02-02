package gateway

type GatewayConfig struct {
	ClusterAddress string

	HTTPAddr string
	GRPCAddr string

	// Логирование
	LogLevel string
}
