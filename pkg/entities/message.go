// Package entities содержит основные сущности брокера сообщений.
package entities

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// MessageOption представляет функциональную опцию для конфигурации Message.
type MessageOption func(*Message)

// Message представляет сообщение в брокере.
type Message struct {
	// ID — уникальный идентификатор сообщения.
	// Используется для дедупликации, трейсинга и acknowledgment.
	ID string

	// RoutingKey определяет в какие пушеры попадет сообщение.
	RoutingKey string

	// Payload содержит полезную нагрузку сообщения.
	Payload []byte
	// Metadata содержит дополнительные метаданные сообщения.
	Metadata map[string]string

	// CreatedAt — время создания сообщения.
	CreatedAt time.Time
	// ScheduledAt — время, когда сообщение должно быть доставлено.
	// Если не задано (zero value), сообщение доставляется немедленно.
	ScheduledAt time.Time
}

// NewMessage создаёт новое сообщение с применением переданных опций.
// По умолчанию ID генерируется автоматически, CreatedAt устанавливается на текущее время.
func NewMessage(options ...MessageOption) *Message {
	message := &Message{
		ID:        generateID(),
		CreatedAt: time.Now(),
	}

	for _, option := range options {
		option(message)
	}

	return message
}

// generateID генерирует уникальный идентификатор.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
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
