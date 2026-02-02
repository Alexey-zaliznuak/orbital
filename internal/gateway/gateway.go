package gateway

import (
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
)

type BaseGateway struct {
	config *gateway.GatewayConfig
}

func (gateway *BaseGateway) Consume(*message.Message) error {
	return nil
}

func (gateway *BaseGateway) GetConfig() *gateway.GatewayConfig {
	return gateway.config
}

func NewBaseGateway() *BaseGateway {
	return &BaseGateway{}
}
