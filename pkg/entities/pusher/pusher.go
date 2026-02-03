package pusher

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/node"
)

// Info описывает метаданные pusher-узла в системе.
type Info struct {
	ID            string
	Type          string // "http", "kafka", "grpc", "nats"
	Address       string
	Status        node.NodeStatus
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

// Pusher отправляет сообщения во внешнюю систему.
type Pusher interface {
	// Push отправляет сообщение.
	Push(msg *message.Message) error
}
