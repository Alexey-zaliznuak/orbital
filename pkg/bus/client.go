// Package bus предоставляет шину сообщений для внутреннего межсервисного общения через NATS.
// Используется внутренними компонентами кластера (gateway и др.) для маршрутизации сообщений.
package bus

import (
	"encoding/json"
	"fmt"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	natsclient "github.com/Alexey-zaliznuak/orbital/pkg/nats"
	"github.com/nats-io/nats.go"
)

// Subject-константы — единственное место определения топологии NATS subjects.
const (
	subjectStoragePrefix = "orbital.storage."
	subjectPusherPrefix  = "orbital.push."
	subjectGateway       = "orbital.gateway"
)

// Client шина сообщений на базе NATS.
type Client struct {
	nats *natsclient.Client
}

// New создаёт новый экземпляр Bus.
func New(nats *natsclient.Client) *Client {
	return &Client{nats: nats}
}

// SendToStorage публикует сообщение в NATS subject orbital.storage.{storageID}.
func (c *Client) SendToStorage(storageID string, msgs []*message.Message) error {
	return c.publish(subjectStoragePrefix+storageID, msgs)
}

// SendToGateway публикует сообщение в NATS subject orbital.gateway
func (c *Client) SendToGateway(msgs []*message.Message) error {
	return c.publish(subjectGateway, msgs)
}

// SendToPusher публикует сообщение в NATS subject orbital.push.{pusherID}.
func (c *Client) SendToPusher(pusherID string, msgs []*message.Message) error {
	return c.publish(subjectPusherPrefix+pusherID, msgs)
}

func (c *Client) NewHandlerOnStorageMessages(storageId string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.nats.Subscribe(subjectStoragePrefix+storageId, handler)
}

func (c *Client) publish(subject string, msgs []*message.Message) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.nats.Publish(subject, data)
}
