package pusher

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/node"
)

// Info описывает метаданные pusher-узла в системе.
type Info struct {
	// ID — уникальный человекочитаемый идентификатор пушера.
	// Используется как часть NATS subject: orbital.push.{ID}.
	// Рекомендуемый формат: "{type}-{name}", например "kafka-prod", "webhook-orders", "grpc-billing".
	// Должен содержать только строчные латинские буквы, цифры и дефис.
	ID            string          `json:"id"`
	Type          string          `json:"type"` // "http", "kafka", "grpc", "nats"
	Address       string          `json:"address"`
	Status        node.NodeStatus `json:"status"`
	RegisteredAt  time.Time       `json:"registered_at"`
	LastHeartbeat time.Time       `json:"last_heartbeat"`
}

// Pusher отправляет сообщения во внешнюю систему.
type Pusher interface {
	// Push отправляет сообщение.
	Push(msg *message.Message) error
}
