package gateway

import (
	"context"
	"fmt"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
	coordinator "github.com/Alexey-zaliznuak/orbital/pkg/sdk/coordinator"
)

type BaseGateway struct {
	config            *gateway.GatewayConfig
	coordinatorClient *coordinator.Client
	storages          []*storage.Info
}

func (g *BaseGateway) Consume(*message.Message) error {
	return nil
}

func (g *BaseGateway) GetConfig() *gateway.GatewayConfig {
	return g.config
}

// GetStorages возвращает список storages.
func (g *BaseGateway) GetStorages() []*storage.Info {
	return g.storages
}

// RefreshStorages обновляет список storages от координатора.
func (g *BaseGateway) RefreshStorages(ctx context.Context) error {
	storages, err := g.coordinatorClient.ListStorages(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh storages: %w", err)
	}
	g.storages = storages
	return nil
}

// NewBaseGateway создаёт новый gateway и получает список storages от координатора.
func NewBaseGateway(ctx context.Context, cfg *gateway.GatewayConfig) (*BaseGateway, error) {
	client := coordinator.NewClient(coordinator.ClientConfig{
		BaseURL: cfg.ClusterAddress,
	})

	storages, err := client.ListStorages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storages from coordinator: %w", err)
	}

	return &BaseGateway{
		config:            cfg,
		coordinatorClient: client,
		storages:          storages,
	}, nil
}
