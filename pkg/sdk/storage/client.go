// Package storage предоставляет HTTP SDK для взаимодействия с storage сервисом.
package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	storageapi "github.com/Alexey-zaliznuak/orbital/pkg/storage/api"
)

const apiPrefix = "/api/v1"

// Client HTTP клиент для взаимодействия со storage сервисом.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ClientConfig конфигурация клиента storage.
type ClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

// NewClient создаёт новый HTTP клиент для storage.
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

// Health проверяет состояние хранилища.
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

// Store сохраняет сообщение в хранилище и возвращает его сохранённую копию.
func (c *Client) Store(ctx context.Context, msg *message.Message) (*message.Message, error) {
	body, err := json.Marshal(storageapi.StoreMessageRequest{
		RoutingKey:      msg.RoutingKey,
		RoutingSettings: msg.RoutingSettings,
		Payload:         msg.Payload,
		Metadata:        msg.Metadata,
		ScheduledAt:     msg.ScheduledAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/messages"), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to store message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.decodeError(resp)
	}

	var result storageapi.MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return messageResponseToMessage(result), nil
}

// FetchReady возвращает до limit готовых к доставке сообщений.
func (c *Client) FetchReady(ctx context.Context, limit int) ([]*message.Message, error) {
	url := c.url("/messages/ready")
	if limit > 0 {
		url += "?limit=" + strconv.Itoa(limit)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ready messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	var result storageapi.FetchReadyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	messages := make([]*message.Message, len(result.Messages))
	for i, m := range result.Messages {
		messages[i] = messageResponseToMessage(m)
	}

	return messages, nil
}

// Acknowledge подтверждает обработку сообщений по их идентификаторам.
func (c *Client) Acknowledge(ctx context.Context, ids []string) error {
	body, err := json.Marshal(storageapi.AcknowledgeRequest{IDs: ids})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/messages/acknowledge"), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to acknowledge messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.decodeError(resp)
	}

	return nil
}

// GetByID возвращает сообщение по идентификатору.
func (c *Client) GetByID(ctx context.Context, id string) (*message.Message, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/messages/"+id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	var result storageapi.MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return messageResponseToMessage(result), nil
}

// Count возвращает общее количество сообщений в хранилище.
func (c *Client) Count(ctx context.Context) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/messages/count"), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, c.decodeError(resp)
	}

	var result storageapi.CountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Count, nil
}

// url формирует полный URL для эндпоинта.
func (c *Client) url(path string) string {
	return c.baseURL + apiPrefix + path
}

// decodeError читает тело ошибки и формирует error.
func (c *Client) decodeError(resp *http.Response) error {
	var errResp storageapi.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil || errResp.Error == "" {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return fmt.Errorf("storage error (status %d): %s", resp.StatusCode, errResp.Error)
}

// messageResponseToMessage преобразует DTO ответа в доменную модель.
func messageResponseToMessage(r storageapi.MessageResponse) *message.Message {
	return message.NewMessage(
		message.WithID(r.ID),
		message.WithRoutingKey(r.RoutingKey),
		message.WithRoutingSettings(r.RoutingSettings),
		message.WithPayload(r.Payload),
		message.WithMetadata(r.Metadata),
		message.WithCreatedAt(r.CreatedAt),
		message.WithScheduledAt(r.ScheduledAt),
	)
}

