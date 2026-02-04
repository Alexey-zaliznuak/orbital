package http

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
)

// ErrorResponse представляет ответ с ошибкой.
// @Description Стандартный ответ при возникновении ошибки.
type ErrorResponse struct {
	// Error содержит описание ошибки.
	Error string `json:"error" example:"invalid request body"`
}

// NewMessageRequest представляет запрос на создание нового сообщения.
// @Description Запрос для отправки сообщения через gateway.
type NewMessageRequest struct {
	// RoutingKey определяет в какие пушеры попадёт сообщение.
	RoutingKey string `json:"routing_key" example:"notifications.email" binding:"required"`

	// Payload содержит полезную нагрузку сообщения.
	Payload []byte `json:"payload" example:"eyJtZXNzYWdlIjogImhlbGxvIn0=" binding:"required"`

	// Metadata содержит дополнительные метаданные сообщения.
	Metadata map[string]string `json:"metadata,omitempty" example:"priority:high,source:api"`

	// ScheduledAt — время, когда сообщение должно быть доставлено.
	// Если не задано (zero value), сообщение доставляется немедленно.
	ScheduledAt time.Time `json:"scheduled_at,omitempty" example:"2024-01-15T10:30:00Z"`
}

// ToMessage преобразует запрос в доменную модель Message.
func (r NewMessageRequest) ToMessage() *message.Message {
	return message.NewMessage(
		message.WithRoutingKey(r.RoutingKey),
		message.WithPayload(r.Payload),
		message.WithMetadata(r.Metadata),
		message.WithScheduledAt(r.ScheduledAt),
	)
}

// NewMessageResponseFromMessage создаёт ответ из доменной модели Message.
func NewMessageResponseFromMessage(m *message.Message) NewMessageResponse {
	return NewMessageResponse{
		ID:          m.ID,
		RoutingKey:  m.RoutingKey,
		Payload:     m.Payload,
		Metadata:    m.Metadata,
		ScheduledAt: m.ScheduledAt,
	}
}

// NewMessageResponse представляет ответ после создания сообщения.
// @Description Ответ с информацией о созданном сообщении.
type NewMessageResponse struct {
	// ID уникальный идентификатор созданного сообщения.
	ID string `json:"id" example:"msg_01HQ3K5X7Y8Z9ABC"`

	// RoutingKey определяет в какие пушеры попадёт сообщение.
	RoutingKey string `json:"routing_key" example:"notifications.email"`

	// Payload содержит полезную нагрузку сообщения.
	Payload []byte `json:"payload" example:"eyJtZXNzYWdlIjogImhlbGxvIn0="`

	// Metadata содержит дополнительные метаданные сообщения.
	Metadata map[string]string `json:"metadata,omitempty" example:"priority:high,source:api"`

	// ScheduledAt — время, когда сообщение должно быть доставлено.
	ScheduledAt time.Time `json:"scheduled_at,omitempty" example:"2024-01-15T10:30:00Z"`
}
