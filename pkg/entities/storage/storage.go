package storage

import (
	"context"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/node"
)

// Info описывает метаданные storage-узла в системе.
type Info struct {
	ID      string
	Address string

	// DelayRange определяет диапазон задержек сообщений, которые хранит это хранилище.
	// Сообщение направляется в это хранилище если:
	//   MinDelay <= (ScheduledAt - now) < MaxDelay
	//
	// Примеры:
	//   Redis:    MinDelay=0,  MaxDelay=1m   (сообщения с задержкой < 1 мин)
	//   Postgres: MinDelay=1m, MaxDelay=1h   (от 1 мин до 1 часа)
	//   S3:       MinDelay=1h, MaxDelay=0    (> 1 часа, 0 = бесконечность)
	MinDelay time.Duration
	MaxDelay time.Duration // 0 означает без верхнего ограничения

	Status        node.NodeStatus
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

// AcceptsDelay проверяет, принимает ли хранилище сообщения с данной задержкой.
func (s *Info) AcceptsDelay(delay time.Duration) bool {
	if delay < s.MinDelay {
		return false
	}
	// MaxDelay == 0 означает без верхнего ограничения
	if s.MaxDelay > 0 && delay >= s.MaxDelay {
		return false
	}
	return true
}

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

// MessageStorage определяет интерфейс хранилища сообщений.
// Каждая реализация (Redis, PostgreSQL, S3) работает со своим диапазоном задержек.
type MessageStorage interface {
	// Store сохраняет сообщение в хранилище.
	Store(ctx context.Context, msg *message.Message) error

	// FetchExpiring возвращает сообщения, у которых ScheduledAt наступает
	// в пределах threshold от текущего времени.
	// Сообщения помечаются как in-flight.
	FetchExpiring(ctx context.Context, threshold time.Duration, limit int) ([]*message.Message, error)

	// FetchReady возвращает сообщения, готовые к отправке (ScheduledAt <= now).
	// Сообщения помечаются как in-flight.
	FetchReady(ctx context.Context, limit int) ([]*message.Message, error)

	// Acknowledge подтверждает успешную обработку сообщения.
	// Удаляет сообщение из хранилища.
	Acknowledge(ctx context.Context, msgID string) error

	// Reject отклоняет сообщение после неудачной обработки.
	// requeue=true — вернуть в очередь для повторной попытки.
	// requeue=false — переместить в dead letter queue или удалить.
	Reject(ctx context.Context, msgID string, requeue bool) error

	// Get возвращает сообщение по ID.
	Get(ctx context.Context, msgID string) (*message.Message, error)

	// Delete удаляет сообщение из хранилища.
	Delete(ctx context.Context, msgID string) error

	// Count возвращает количество сообщений в хранилище.
	Count(ctx context.Context) (int64, error)

	// Close закрывает соединение с хранилищем.
	Close() error
}
