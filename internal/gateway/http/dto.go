package http

import (
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// Message

type NewMessageRequest struct {
	// RoutingKey определяет в какие пушеры попадет сообщение.
	RoutingKey string `json:"routing_key"`

	// Payload содержит полезную нагрузку сообщения.
	Payload []byte `json:"payload"`
	// Metadata содержит дополнительные метаданные сообщения.
	Metadata map[string]string `json:"metadata,omitempty"`

	// ScheduledAt — время, когда сообщение должно быть доставлено.
	// Если не задано (zero value), сообщение доставляется немедленно.
	ScheduledAt time.Time `json:"scheduled_at,omitzero"`
}

type NewMessageResponse struct {
	ID string `json:"id"`

	// RoutingKey определяет в какие пушеры попадет сообщение.
	RoutingKey string `json:"routing_key"`

	// Payload содержит полезную нагрузку сообщения.
	Payload []byte `json:"payload"`

	// Metadata содержит дополнительные метаданные сообщения.
	Metadata map[string]string `json:"metadata,omitempty"`

	// ScheduledAt — время, когда сообщение должно быть доставлено.
	// Если не задано (zero value), сообщение доставляется немедленно.
	ScheduledAt time.Time `json:"scheduled_at,omitzero"`
}

type StorageResponse struct {
	ID            string `json:"id"`
	Address       string `json:"address"`
	MinDelay      string `json:"min_delay"`
	MaxDelay      string `json:"max_delay"`
	Status        string `json:"status"`
	RegisteredAt  string `json:"registered_at"`
	LastHeartbeat string `json:"last_heartbeat"`
}

type ListStoragesResponse []StorageResponse
