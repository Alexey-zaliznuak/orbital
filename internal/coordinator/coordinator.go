package coordinator

import (
	"context"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
)

type BaseCoordinator struct {
	storage           coordinator.CoordinatorStorage
	clusterConfig     *coordinator.ClusterConfig
	coordinatorConfig *coordinator.CoordinatorConfig
}

func (c *BaseCoordinator) GetStorage() coordinator.CoordinatorStorage {
	return c.storage
}

func (c *BaseCoordinator) GetClusterConfig() *coordinator.ClusterConfig {
	return c.clusterConfig
}

func (c *BaseCoordinator) GetCoordinatorConfig() *coordinator.CoordinatorConfig {
	return c.coordinatorConfig
}

func NewBaseCoordinator(ctx context.Context, storage coordinator.CoordinatorStorage, coordinatorConfig *coordinator.CoordinatorConfig, clusterConfig *coordinator.ClusterConfig) (*BaseCoordinator, error) {
	coordinator := &BaseCoordinator{
		storage:           storage,
		coordinatorConfig: coordinatorConfig,
		clusterConfig:     clusterConfig,
	}
	return coordinator, nil
}
