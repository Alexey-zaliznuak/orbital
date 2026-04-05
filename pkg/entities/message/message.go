package message

import (
	"time"

	"github.com/google/uuid"
)

// MessageOption представляет функциональную опцию для конфигурации Message.
type MessageOption func(*Message)

// Message представляет сообщение в брокере.
type Message struct {
	ID string `json:"id"`

	// Позволяет routing rule определить, в какой pusher отправить сообщение.
	RoutingKey string `json:"routing_key"`

	// Может быть использован pusher для дополнительной параметризации.
	// Example:  http headers, query params, etc.
	RoutingSettings map[string]string `json:"routing_settings"`

	Payload []byte `json:"payload"`

	Metadata map[string]string `json:"metadata"`

	CreatedAt time.Time `json:"created_at"`

	// Если не задано (zero value), сообщение доставляется немедленно.
	ScheduledAt time.Time `json:"scheduled_at"`
}

// NewMessage создаёт новое сообщение с применением переданных опций.
// По умолчанию ID генерируется автоматически, CreatedAt устанавливается на текущее время.
func NewMessage(options ...MessageOption) *Message {
	message := &Message{
		ID:        GenerateID(),
		CreatedAt: time.Now(),
	}

	for _, option := range options {
		option(message)
	}

	return message
}

// GenerateID генерирует временно-упорядоченный идентификатор UUIDv6.
// Используется как единый стандарт генерации ID во всех слоях системы.
// При недоступности источника энтропии выполняется fallback на UUIDv4.
func GenerateID() string {
	id, err := uuid.NewV6()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}

// WithID устанавливает идентификатор сообщения.
// Используйте для восстановления сообщения из хранилища.
func WithID(id string) MessageOption {
	return func(m *Message) {
		m.ID = id
	}
}

// WithRoutingKey устанавливает ключ маршрутизации сообщения.
func WithRoutingKey(key string) MessageOption {
	return func(m *Message) {
		m.RoutingKey = key
	}
}

// WithRoutingSettings устанавливает настройки маршрутизации сообщения.
func WithRoutingSettings(settings map[string]string) MessageOption {
	return func(m *Message) {
		m.RoutingSettings = settings
	}
}

// WithPayload устанавливает полезную нагрузку сообщения.
func WithPayload(payload []byte) MessageOption {
	return func(m *Message) {
		m.Payload = payload
	}
}

// WithMetadata устанавливает метаданные сообщения целиком.
func WithMetadata(metadata map[string]string) MessageOption {
	return func(m *Message) {
		m.Metadata = metadata
	}
}

// WithMetadataValue добавляет одну пару ключ-значение в метаданные.
// Если метаданные ещё не инициализированы, создаёт новую map.
func WithMetadataValue(key, value string) MessageOption {
	return func(m *Message) {
		if m.Metadata == nil {
			m.Metadata = make(map[string]string)
		}
		m.Metadata[key] = value
	}
}

// WithCreatedAt переопределяет время создания сообщения.
func WithCreatedAt(t time.Time) MessageOption {
	return func(m *Message) {
		m.CreatedAt = t
	}
}

// WithScheduledAt устанавливает точное время доставки сообщения.
func WithScheduledAt(t time.Time) MessageOption {
	return func(m *Message) {
		m.ScheduledAt = t
	}
}

// WithDelay устанавливает задержку доставки относительно текущего времени.
func WithDelay(d time.Duration) MessageOption {
	return func(m *Message) {
		m.ScheduledAt = time.Now().Add(d)
	}
}
