package coordinator

import "context"

type Coordinator interface {
	GetStorage() CoordinatorStorage
	GetClusterConfig() *ClusterConfig
	GetCoordinatorConfig() *CoordinatorConfig
}

type BaseCoordinator struct {
	storage           CoordinatorStorage
	clusterConfig     *ClusterConfig
	coordinatorConfig *CoordinatorConfig
}

func (c *BaseCoordinator) GetStorage() CoordinatorStorage {
	return c.storage
}

func (c *BaseCoordinator) GetClusterConfig() *ClusterConfig {
	return c.clusterConfig
}

func (c *BaseCoordinator) GetCoordinatorConfig() *CoordinatorConfig {
	return c.coordinatorConfig
}

func NewBaseCoordinator(ctx context.Context, storage CoordinatorStorage, coordinatorConfig *CoordinatorConfig, clusterConfig *ClusterConfig) (*BaseCoordinator, error) {
	coordinator := &BaseCoordinator{
		storage:           storage,
		coordinatorConfig: coordinatorConfig,
		clusterConfig:     clusterConfig,
	}
	return coordinator, nil
}
