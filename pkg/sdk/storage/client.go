// Package storage предоставляет SDK для отправки сообщений в storage через NATS.
package storage

import (
	"encoding/json"
	"fmt"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/nats"
)

const storagesSubjectPrefix = "orbital.storage."

// Client представляет SDK для отправки сообщений в storage через NATS.
type Client struct {
	nats *natsclient.Client
}

// New создаёт новый экземпляр Client.
func New(nats *natsclient.Client) *Client {
	return &Client{nats: nats}
}

// SendMessage публикует сообщение в NATS subject orbital.storage.{storageID}.
func (c *Client) SendMessage(storageID string, msg *message.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	subject := storagesSubjectPrefix + storageID
	return c.nats.Publish(subject, data)
}
