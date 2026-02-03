package coordinator

import (
	"context"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/pusher"
	routingrule "github.com/Alexey-zaliznuak/orbital/pkg/entities/routing_rule"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
	"github.com/google/uuid"
)

type CoordinatorStorage interface {
	// === Coordinator Nodes ===
	CreateNode(ctx context.Context, node *Node) error
	GetNode(ctx context.Context, nodeID uuid.UUID) (*Node, error)
	ListNodes(ctx context.Context) ([]*Node, error)
	UpdateNodeHeartbeat(ctx context.Context, nodeID uuid.UUID) error
	DeleteNode(ctx context.Context, nodeID uuid.UUID) error

	// === Gateways ===
	RegisterGateway(ctx context.Context, gw *gateway.Info) error
	GetGateway(ctx context.Context, gatewayID string) (*gateway.Info, error)
	ListGateways(ctx context.Context) ([]*gateway.Info, error)
	UpdateGatewayHeartbeat(ctx context.Context, gatewayID string) error
	UnregisterGateway(ctx context.Context, gatewayID string) error

	// === Storages ===
	RegisterStorage(ctx context.Context, st *storage.Info) error
	GetStorage(ctx context.Context, storageID string) (*storage.Info, error)
	ListStorages(ctx context.Context) ([]*storage.Info, error)
	UpdateStorageHeartbeat(ctx context.Context, storageID string) error
	UnregisterStorage(ctx context.Context, storageID string) error

	// === Pushers ===
	RegisterPusher(ctx context.Context, p *pusher.Info) error
	GetPusher(ctx context.Context, pusherID string) (*pusher.Info, error)
	ListPushers(ctx context.Context) ([]*pusher.Info, error)
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
