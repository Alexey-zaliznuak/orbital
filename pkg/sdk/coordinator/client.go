// Package coordinator предоставляет SDK для взаимодействия с координатором кластера.
package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	coordinatorapi "github.com/Alexey-zaliznuak/orbital/pkg/coordinator/api"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/pusher"
	routingrule "github.com/Alexey-zaliznuak/orbital/pkg/entities/routing_rule"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
)

const apiPrefix = "/api/v1"

// Client HTTP клиент для взаимодействия с координатором.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ClientConfig конфигурация клиента координатора.
type ClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

// NewClient создаёт новый клиент для координатора.
func NewClient(cfg ClientConfig) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ListStorages получает список всех storages от координатора.
func (c *Client) ListStorages(ctx context.Context) ([]*storage.Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+apiPrefix+"/storages", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var responses []coordinatorapi.StorageResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	storages := make([]*storage.Info, 0, len(responses))
	for i := range responses {
		info, err := coordinatorapi.ParseStorageResponse(&responses[i])
		if err != nil {
			continue // skip invalid entries
		}
		storages = append(storages, info)
	}

	return storages, nil
}

// === Cluster Config ===

// GetClusterConfig получает конфигурацию кластера от координатора.
func (c *Client) GetClusterConfig(ctx context.Context) (*coordinator.ClusterConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+apiPrefix+"/cluster-config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cluster config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var cfg coordinator.ClusterConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &cfg, nil
}

// === Pushers ===

// ListPushers получает список всех pushers от координатора.
func (c *Client) ListPushers(ctx context.Context) ([]*pusher.Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+apiPrefix+"/pushers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pushers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var pushers []*pusher.Info
	if err := json.NewDecoder(resp.Body).Decode(&pushers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return pushers, nil
}

// === Routing Rules ===

// ListRoutingRules получает список всех routing rules от координатора.
func (c *Client) ListRoutingRules(ctx context.Context) ([]*routingrule.RoutingRule, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+apiPrefix+"/routing-rules", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routing rules: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rules []*routingrule.RoutingRule
	if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Компилируем regex для правил
	for _, rule := range rules {
		rule.CompileRegex()
	}

	return rules, nil
}
