// Package natsclient предоставляет обёртку над NATS с поддержкой
// JetStream и интеграцией с логгером проекта.
package natsclient

import (
	"fmt"

	"github.com/Alexey-zaliznuak/orbital/pkg/logger"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Client обёртка над NATS с JetStream и логированием.
type Client struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// Config конфигурация NATS клиента.
type Config struct {
	// URL адрес NATS сервера (например, "nats://localhost:4222").
	URL string
	// Name имя клиента для идентификации в NATS (опционально).
	Name string
}

// New создаёт новое подключение к NATS и инициализирует JetStream.
func New(cfg Config) (*Client, error) {
	opts := []nats.Option{
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			logger.Log.Error("NATS async error", zap.Error(err))
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logger.Log.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Log.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			logger.Log.Info("NATS connection closed")
		}),
	}

	if cfg.Name != "" {
		opts = append(opts, nats.Name(cfg.Name))
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to init JetStream: %w", err)
	}

	logger.Log.Info("NATS connected",
		zap.String("url", conn.ConnectedUrl()),
		zap.String("name", cfg.Name),
	)

	return &Client{conn: conn, js: js}, nil
}

// Conn возвращает базовое NATS подключение.
func (c *Client) Conn() *nats.Conn {
	return c.conn
}

// JetStream возвращает JetStream контекст.
func (c *Client) JetStream() nats.JetStreamContext {
	return c.js
}

// Publish публикует сообщение в NATS subject через JetStream.
func (c *Client) Publish(subject string, data []byte) error {
	_, err := c.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish to %s: %w", subject, err)
	}
	return nil
}

// Close закрывает соединение с NATS.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
