package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/bus"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/pusher"
	routingrule "github.com/Alexey-zaliznuak/orbital/pkg/entities/routing_rule"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
	"github.com/Alexey-zaliznuak/orbital/pkg/logger"
	natsclient "github.com/Alexey-zaliznuak/orbital/pkg/nats"
	"github.com/Alexey-zaliznuak/orbital/pkg/sdk/coordinator"
	"go.uber.org/zap"
)

type BaseGateway struct {
	config            *gateway.GatewayConfig
	coordinatorClient *coordinator.Client
	natsClient        *natsclient.Client
	bus               *bus.Client

	storages   []*storage.Info
	storagesMu sync.RWMutex

	pushers   []*pusher.Info
	pushersMu sync.RWMutex

	routingRules   []*routingrule.RoutingRule
	routingRulesMu sync.RWMutex

	refreshPeriod time.Duration

	minDelayForSaveInStorage time.Duration
}

func (g *BaseGateway) Consume(message *message.Message) error {
	delay := time.Until(message.ScheduledAt)

	if delay <= g.minDelayForSaveInStorage {
		return g.sendToPusher(message)
	}

	return g.sendToStorage(message)
}

func (g *BaseGateway) sendToStorage(msg *message.Message) error {
	storages := g.GetStorages()
	delay := time.Until(msg.ScheduledAt)

	for _, storage := range storages {
		if storage.AcceptsDelay(delay) {
			return g.bus.SendToStorage(storage.ID, []*message.Message{msg})
		}
	}

	logger.Log.Warn(
		"No storages for saving message, sending to pusher",
		zap.String("id", msg.ID),
		zap.String("key", msg.RoutingKey),
		zap.Time("scheduledAt", msg.ScheduledAt),
	)

	return g.sendToPusher(msg)
}

func (g *BaseGateway) sendToPusher(msg *message.Message) error {
	rules := g.GetRoutingRules()

	for _, rule := range rules {
		if rule.Match(msg.RoutingKey) {
			return g.bus.SendToPusher(rule.PusherID, []*message.Message{msg})
		}
	}

	logger.Log.Warn(
		"No pusher for sending message, message will be dropped",
		zap.String("id", msg.ID),
		zap.String("key", msg.RoutingKey),
	)

	return nil
}

func (g *BaseGateway) Start(ctx context.Context) {
	go g.runRefreshLoop(ctx)
}

func (g *BaseGateway) runRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(g.refreshPeriod)
	defer ticker.Stop()

	g.refreshAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.refreshAll(ctx)
		}
	}
}

func (g *BaseGateway) refreshAll(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() { defer wg.Done(); g.RefreshStorages(ctx) }()
	go func() { defer wg.Done(); g.RefreshRoutingRules(ctx) }()
	go func() { defer wg.Done(); g.RefreshPushers(ctx) }()

	wg.Wait()
}

func (g *BaseGateway) GetConfig() *gateway.GatewayConfig {
	return g.config
}

// GetStorages возвращает список storages.
func (g *BaseGateway) GetStorages() []*storage.Info {
	g.storagesMu.RLock()
	defer g.storagesMu.RUnlock()
	return g.storages
}

// RefreshStorages обновляет список storages от координатора.
func (g *BaseGateway) RefreshStorages(ctx context.Context) error {
	storages, err := g.coordinatorClient.ListStorages(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh storages: %w", err)
	}

	g.storagesMu.Lock()
	g.storages = storages
	g.storagesMu.Unlock()

	logger.GetFromContext(ctx).Debug("Storages info refreshed")
	return nil
}

// GetPushers возвращает список pushers.
func (g *BaseGateway) GetPushers() []*pusher.Info {
	g.pushersMu.RLock()
	defer g.pushersMu.RUnlock()
	return g.pushers
}

// RefreshPushers обновляет список pushers от координатора.
func (g *BaseGateway) RefreshPushers(ctx context.Context) error {
	pushers, err := g.coordinatorClient.ListPushers(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh pushers: %w", err)
	}

	g.pushersMu.Lock()
	g.pushers = pushers
	g.pushersMu.Unlock()

	logger.GetFromContext(ctx).Debug("Pushers info refreshed")
	return nil
}

// GetRoutingRules возвращает список routing rules.
func (g *BaseGateway) GetRoutingRules() []*routingrule.RoutingRule {
	g.routingRulesMu.RLock()
	defer g.routingRulesMu.RUnlock()
	return g.routingRules
}

// RefreshRoutingRules обновляет список routing rules от координатора.
func (g *BaseGateway) RefreshRoutingRules(ctx context.Context) error {
	rules, err := g.coordinatorClient.ListRoutingRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh routing rules: %w", err)
	}

	g.routingRulesMu.Lock()
	g.routingRules = rules
	g.routingRulesMu.Unlock()

	logger.GetFromContext(ctx).Debug("Routing rules info refreshed")
	return nil
}

// NewBaseGateway создаёт новый gateway.
// Получает адрес NATS из координатора и подключается к нему.
func NewBaseGateway(ctx context.Context, cfg *gateway.GatewayConfig) (*BaseGateway, error) {
	coordinatorClient := coordinator.NewClient(coordinator.ClientConfig{
		BaseURL: cfg.ClusterAddress,
	})

	clusterCfg, err := coordinatorClient.GetClusterConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config: %w", err)
	}

	nc, err := natsclient.New(natsclient.Config{
		URL:  clusterCfg.NatsAddress,
		Name: "gateway",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	g := &BaseGateway{
		config:                   cfg,
		coordinatorClient:        coordinatorClient,
		natsClient:               nc,
		bus:                      bus.New(nc),
		storages:                 make([]*storage.Info, 0),
		pushers:                  make([]*pusher.Info, 0),
		routingRules:             make([]*routingrule.RoutingRule, 0),
		minDelayForSaveInStorage: time.Millisecond * 10, // TODO перенести в конфиг
		refreshPeriod:            time.Second * 10,      // TODO перенести в конфиг
	}

	return g, nil
}
