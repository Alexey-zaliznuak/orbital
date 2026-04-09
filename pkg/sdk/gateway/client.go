// Package gateway предоставляет HTTP SDK для взаимодействия с gateway сервисом.
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	gatewayapi "github.com/Alexey-zaliznuak/orbital/pkg/sdk/gateway/api"
)

const apiPrefix = "/api/v1"

// Client HTTP клиент для взаимодействия с gateway сервисом.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ClientConfig конфигурация клиента gateway.
type ClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

// NewClient создаёт новый HTTP клиент для gateway.
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

// Health проверяет состояние gateway.
func (c *Client) Health(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/health"), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result["status"], nil
}

// Send отправляет сообщение в gateway и возвращает созданное сообщение.
func (c *Client) Send(ctx context.Context, msg *message.Message) (*message.Message, error) {
	body, err := json.Marshal(gatewayapi.NewMessageRequest{
		RoutingKey:  msg.RoutingKey,
		Payload:     msg.Payload,
		Metadata:    msg.Metadata,
		ScheduledAt: msg.ScheduledAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/message"), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.decodeError(resp)
	}

	var result gatewayapi.NewMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return newMessageResponseToMessage(result), nil
}

// url формирует полный URL для эндпоинта.
func (c *Client) url(path string) string {
	return c.baseURL + apiPrefix + path
}

// decodeError читает тело ошибки и формирует error.
func (c *Client) decodeError(resp *http.Response) error {
	var errResp gatewayapi.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil || errResp.Error == "" {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return fmt.Errorf("gateway error (status %d): %s", resp.StatusCode, errResp.Error)
}

// newMessageResponseToMessage преобразует DTO ответа в доменную модель.
func newMessageResponseToMessage(r gatewayapi.NewMessageResponse) *message.Message {
	return message.NewMessage(
		message.WithID(r.ID),
		message.WithRoutingKey(r.RoutingKey),
		message.WithPayload(r.Payload),
		message.WithMetadata(r.Metadata),
		message.WithScheduledAt(r.ScheduledAt),
	)
}
