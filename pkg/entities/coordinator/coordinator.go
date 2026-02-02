package coordinator

type Coordinator interface {
	GetStorage() CoordinatorStorage
	GetClusterConfig() *ClusterConfig
	GetCoordinatorConfig() *CoordinatorConfig
}
