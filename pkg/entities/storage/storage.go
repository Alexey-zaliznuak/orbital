package storage

import (
	"context"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
)

// MessageStatus определяет статус сообщения в хранилище.
type MessageStatus int

const (
	// MessageStatusPending — сообщение ожидает обработки.
	MessageStatusPending MessageStatus = iota
	// MessageStatusInFlight — сообщение в процессе доставки.
	MessageStatusInFlight
	// MessageStatusDelivered — сообщение доставлено.
	MessageStatusDelivered
	// MessageStatusFailed — доставка не удалась.
	MessageStatusFailed
)

// StoredMessage представляет сообщение в хранилище с дополнительными метаданными.
type StoredMessage struct {
	// Message — само сообщение.
	*message.Message
	// Status — текущий статус сообщения.
	Status MessageStatus
	// Attempts — количество попыток доставки.
	Attempts int
	// LastAttemptAt — время последней попытки.
	LastAttemptAt time.Time
}

// MessageStorage определяет интерфейс хранилища сообщений.
// Каждая реализация (Redis, PostgreSQL, S3) работает со своим диапазоном задержек.
type MessageStorage interface {
	// Store сохраняет сообщение в хранилище.
	Store(ctx context.Context, msg *message.Message) error

	// FetchExpiring возвращает сообщения, у которых ScheduledAt наступает
	// в пределах threshold от текущего времени.
	// Сообщения помечаются как in-flight.
	FetchExpiring(ctx context.Context, threshold time.Duration, limit int) ([]*StoredMessage, error)

	// FetchReady возвращает сообщения, готовые к отправке (ScheduledAt <= now).
	// Сообщения помечаются как in-flight.
	FetchReady(ctx context.Context, limit int) ([]*StoredMessage, error)

	// Acknowledge подтверждает успешную обработку сообщения.
	// Удаляет сообщение из хранилища.
	Acknowledge(ctx context.Context, msgID string) error

	// Reject отклоняет сообщение после неудачной обработки.
	// requeue=true — вернуть в очередь для повторной попытки.
	// requeue=false — переместить в dead letter queue или удалить.
	Reject(ctx context.Context, msgID string, requeue bool) error

	// Get возвращает сообщение по ID.
	Get(ctx context.Context, msgID string) (*StoredMessage, error)

	// Delete удаляет сообщение из хранилища.
	Delete(ctx context.Context, msgID string) error

	// Count возвращает количество сообщений в хранилище.
	Count(ctx context.Context) (int64, error)

	// Close закрывает соединение с хранилищем.
	Close() error
}
