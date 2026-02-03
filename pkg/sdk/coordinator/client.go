// Package coordinator предоставляет SDK для взаимодействия с координатором кластера.
package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/node"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
)

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

// storageResponse DTO для ответа от координатора.
type storageResponse struct {
	ID            string `json:"id"`
	Address       string `json:"address"`
	MinDelay      string `json:"min_delay"`
	MaxDelay      string `json:"max_delay"`
	Status        string `json:"status"`
	RegisteredAt  string `json:"registered_at"`
	LastHeartbeat string `json:"last_heartbeat"`
}

// ListStorages получает список всех storages от координатора.
func (c *Client) ListStorages(ctx context.Context) ([]*storage.Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/storages", nil)
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

	var responses []storageResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	storages := make([]*storage.Info, 0, len(responses))
	for _, r := range responses {
		info, err := parseStorageResponse(&r)
		if err != nil {
			continue // skip invalid entries
		}
		storages = append(storages, info)
	}

	return storages, nil
}

func parseStorageResponse(r *storageResponse) (*storage.Info, error) {
	minDelay, err := time.ParseDuration(r.MinDelay)
	if err != nil {
		return nil, fmt.Errorf("invalid min_delay: %w", err)
	}

	var maxDelay time.Duration
	if r.MaxDelay != "unlimited" && r.MaxDelay != "" {
		maxDelay, err = time.ParseDuration(r.MaxDelay)
		if err != nil {
			return nil, fmt.Errorf("invalid max_delay: %w", err)
		}
	}

	registeredAt, _ := time.Parse(time.RFC3339, r.RegisteredAt)
	lastHeartbeat, _ := time.Parse(time.RFC3339, r.LastHeartbeat)

	status := node.NodeStatusActive
	if r.Status != "Active" {
		status = node.NodeStatusRemoved
	}

	return &storage.Info{
		ID:            r.ID,
		Address:       r.Address,
		MinDelay:      minDelay,
		MaxDelay:      maxDelay,
		Status:        status,
		RegisteredAt:  registeredAt,
		LastHeartbeat: lastHeartbeat,
	}, nil
}
