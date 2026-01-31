package coordinator

import (
	"context"
	"time"

	routingrule "github.com/Alexey-zaliznuak/orbital/pkg/entities/routing_rule"
	"github.com/google/uuid"
)

type GatewayInfo struct {
	ID            string
	Address       string
	Status        NodeStatus
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

type StorageInfo struct {
	ID      string
	Address string

	// DelayRange определяет диапазон задержек сообщений, которые хранит это хранилище.
	// Сообщение направляется в это хранилище если:
	//   MinDelay <= (ScheduledAt - now) < MaxDelay
	//
	// Примеры:
	//   Redis:    MinDelay=0,  MaxDelay=1m   (сообщения с задержкой < 1 мин)
	//   Postgres: MinDelay=1m, MaxDelay=1h   (от 1 мин до 1 часа)
	//   S3:       MinDelay=1h, MaxDelay=0    (> 1 часа, 0 = бесконечность)
	MinDelay time.Duration
	MaxDelay time.Duration // 0 означает без верхнего ограничения

	Status        NodeStatus
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

// AcceptsDelay проверяет, принимает ли хранилище сообщения с данной задержкой.
func (s *StorageInfo) AcceptsDelay(delay time.Duration) bool {
	if delay < s.MinDelay {
		return false
	}
	// MaxDelay == 0 означает без верхнего ограничения
	if s.MaxDelay > 0 && delay >= s.MaxDelay {
		return false
	}
	return true
}

type PusherInfo struct {
	ID            string
	Type          string // "http", "kafka", "grpc", "nats"
	Address       string
	Status        NodeStatus
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

type CoordinatorStorage interface {
	// === Coordinator Nodes ===
	CreateNode(ctx context.Context, node *Node) error
	GetNode(ctx context.Context, nodeID uuid.UUID) (*Node, error)
	ListNodes(ctx context.Context) ([]*Node, error)
	UpdateNodeHeartbeat(ctx context.Context, nodeID uuid.UUID) error
	DeleteNode(ctx context.Context, nodeID uuid.UUID) error

	// === Gateways ===
	RegisterGateway(ctx context.Context, gateway *GatewayInfo) error
	GetGateway(ctx context.Context, gatewayID string) (*GatewayInfo, error)
	ListGateways(ctx context.Context) ([]*GatewayInfo, error)
	UpdateGatewayHeartbeat(ctx context.Context, gatewayID string) error
	UnregisterGateway(ctx context.Context, gatewayID string) error

	// === Storages ===
	RegisterStorage(ctx context.Context, storage *StorageInfo) error
	GetStorage(ctx context.Context, storageID string) (*StorageInfo, error)
	ListStorages(ctx context.Context) ([]*StorageInfo, error)
	UpdateStorageHeartbeat(ctx context.Context, storageID string) error
	UnregisterStorage(ctx context.Context, storageID string) error

	// === Pushers ===
	RegisterPusher(ctx context.Context, pusher *PusherInfo) error
	GetPusher(ctx context.Context, pusherID string) (*PusherInfo, error)
	ListPushers(ctx context.Context) ([]*PusherInfo, error)
	UpdatePusherHeartbeat(ctx context.Context, pusherID string) error
	UnregisterPusher(ctx context.Context, pusherID string) error

	// === Routing Rules ===
	CreateRoutingRule(ctx context.Context, rule *routingrule.RoutingRule) error
	GetRoutingRule(ctx context.Context, ruleID string) (*routingrule.RoutingRule, error)
	ListRoutingRules(ctx context.Context) ([]*routingrule.RoutingRule, error)
	UpdateRoutingRule(ctx context.Context, rule *routingrule.RoutingRule) error
	DeleteRoutingRule(ctx context.Context, ruleID string) error

	// === Coordinator Config ===
	GetCoordinatorConfig(ctx context.Context) (*CoordinatorConfig, error)
	SetCoordinatorConfig(ctx context.Context, config *CoordinatorConfig) error

	// === Cluster Config ===
	GetClusterConfig(ctx context.Context) (*ClusterConfig, error)
	SetClusterConfig(ctx context.Context, config *ClusterConfig) error
}
