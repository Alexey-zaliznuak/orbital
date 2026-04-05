package storageapi

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
)

// ErrorResponse стандартный ответ при ошибке.
type ErrorResponse struct {
	Error string `json:"error"`
}

// StoreMessageRequest запрос на сохранение сообщения.
type StoreMessageRequest struct {
	RoutingKey      string            `json:"routing_key"`
	RoutingSettings map[string]string `json:"routing_settings,omitempty"`
	Payload         []byte            `json:"payload"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	ScheduledAt     time.Time         `json:"scheduled_at,omitempty"`
}

// ToMessage преобразует запрос в доменную модель Message.
func (r StoreMessageRequest) ToMessage() *message.Message {
	return message.NewMessage(
		message.WithRoutingKey(r.RoutingKey),
		message.WithRoutingSettings(r.RoutingSettings),
		message.WithPayload(r.Payload),
		message.WithMetadata(r.Metadata),
		message.WithScheduledAt(r.ScheduledAt),
	)
}

// MessageResponse представляет сообщение в ответе API.
type MessageResponse struct {
	ID              string            `json:"id"`
	RoutingKey      string            `json:"routing_key"`
	RoutingSettings map[string]string `json:"routing_settings,omitempty"`
	Payload         []byte            `json:"payload"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	ScheduledAt     time.Time         `json:"scheduled_at,omitempty"`
}

// MessageResponseFromMessage создаёт ответ из доменной модели Message.
func MessageResponseFromMessage(m *message.Message) MessageResponse {
	return MessageResponse{
		ID:              m.ID,
		RoutingKey:      m.RoutingKey,
		RoutingSettings: m.RoutingSettings,
		Payload:         m.Payload,
		Metadata:        m.Metadata,
		CreatedAt:       m.CreatedAt,
		ScheduledAt:     m.ScheduledAt,
	}
}

// FetchReadyResponse ответ на запрос готовых к доставке сообщений.
type FetchReadyResponse struct {
	Messages []MessageResponse `json:"messages"`
}

// AcknowledgeRequest запрос на подтверждение обработки сообщений.
type AcknowledgeRequest struct {
	IDs []string `json:"ids"`
}

// CountResponse ответ с количеством сообщений в хранилище.
type CountResponse struct {
	Count int64 `json:"count"`
}
