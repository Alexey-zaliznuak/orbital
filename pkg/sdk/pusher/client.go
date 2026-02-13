// Package pusher предоставляет SDK для отправки сообщений в pushers через NATS.
package pusher

import (
	"encoding/json"
	"fmt"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	natsclient "github.com/Alexey-zaliznuak/orbital/pkg/nats"
)

const subjectPrefix = "orbital.push."

// Client представляет SDK для отправки сообщений в pushers через NATS.
type Client struct {
	nats *natsclient.Client
}

// New создаёт новый экземпляр Client.
func New(nats *natsclient.Client) *Client {
	return &Client{nats: nats}
}

// SendMessage публикует сообщение в NATS subject orbital.push.{pusherID}.
func (c *Client) SendMessage(pusherID string, msg *message.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	subject := subjectPrefix + pusherID
	return c.nats.Publish(subject, data)
}
