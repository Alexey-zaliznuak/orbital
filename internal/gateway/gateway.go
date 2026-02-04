package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
	"github.com/Alexey-zaliznuak/orbital/pkg/logger"
	coordinator "github.com/Alexey-zaliznuak/orbital/pkg/sdk/coordinator"
)

type BaseGateway struct {
	config            *gateway.GatewayConfig
	coordinatorClient *coordinator.Client
	httpClient        *http.Client

	storages            []*storage.Info
	storagesLastRefresh time.Time
	storagesRefreshMu   sync.RWMutex

	storagesRefreshPeriod time.Duration

	minDelayForSaveInStorage time.Duration
}

func (g *BaseGateway) Consume(message *message.Message) error {
	delay := time.Until(message.ScheduledAt)

	if delay <= g.minDelayForSaveInStorage {
		return g.sendToPusher(message)
	}

	return g.sendToStorage(message)
}

func (g *BaseGateway) sendToStorage(message *message.Message) error {
	storages := g.GetStorages()
	delay := time.Until(message.ScheduledAt)

	for _, storage := range storages {
		if storage.MinDelay <= delay && delay <= storage.MaxDelay {
			
		}
	}

	return nil
}

func (g *BaseGateway) sendToPusher(message *message.Message) error {
	return nil
}

func (g *BaseGateway) Start(ctx context.Context) {
	go g.runStoragesRefreshLoop(ctx)
}

func (g *BaseGateway) runStoragesRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(g.storagesRefreshPeriod)
	defer ticker.Stop()

	g.RefreshStorages(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.RefreshStorages(ctx)
		}
	}
}

func (g *BaseGateway) GetConfig() *gateway.GatewayConfig {
	return g.config
}

// GetStorages возвращает список storages.
func (g *BaseGateway) GetStorages() []*storage.Info {
	g.storagesRefreshMu.RLock()
	defer g.storagesRefreshMu.RUnlock()
	return g.storages
}

// RefreshStorages обновляет список storages от координатора.
func (g *BaseGateway) RefreshStorages(ctx context.Context) error {
	g.storagesRefreshMu.Lock()
	defer g.storagesRefreshMu.Unlock()

	storages, err := g.coordinatorClient.ListStorages(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh storages: %w", err)
	}

	g.storages = storages

	logger.GetFromContext(ctx).Debug("Storages info refreshed")
	return nil
}

// NewBaseGateway создаёт новый gateway и получает список storages от координатора.
func NewBaseGateway(ctx context.Context, cfg *gateway.GatewayConfig) (*BaseGateway, error) {
	client := coordinator.NewClient(coordinator.ClientConfig{
		BaseURL: cfg.ClusterAddress,
	})

	g := &BaseGateway{
		config:                   cfg,
		coordinatorClient:        client,
		storages:                 make([]*storage.Info, 10),
		storagesLastRefresh:      time.UnixMicro(0),
		storagesRefreshPeriod:    time.Second * 10,      // TODO перенести в конфиг
		minDelayForSaveInStorage: time.Millisecond * 10, // TODO перенести в конфиг
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // TODO перенести в конфиг
		},
	}

	return g, nil
}
