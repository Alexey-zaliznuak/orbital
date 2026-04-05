package storage

import (
	"context"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/node"
)

type StorageHealth string

const (
	// Хранилище работает нормально.
	StorageHealthOK StorageHealth = "ok"
	// Хранилище недоступно.
	StorageHealthDisconnect StorageHealth = "disconnect"
	// Произошла ошибка при проверке состояния хранилища.
	StorageHealthError StorageHealth = "error"
)

// Info описывает метаданные storage-узла в системе.
type Info struct {
	// ID — уникальный человекочитаемый идентификатор хранилища.
	// Используется как часть NATS subject: orbital.storage.{ID}.
	// Формат: произвольный, предпочтительно "{tier}-{level}", например "hot-l1", "warm-l1", "cold-l1".
	// Должен содержать только строчные латинские буквы, цифры и дефис.
	ID      string
	Address string

	// DelayRange определяет диапазон задержек сообщений, которые хранит это хранилище.
	// Сообщение направляется в это хранилище если:
	//   MinDelay <= (ScheduledAt - now) < MaxDelay
	//
	// Примеры:
	//   Redis:    MinDelay=0,  MaxDelay=1m   (сообщения с задержкой < 1 мин)
	//   Postgres: MinDelay=1m, MaxDelay=1h   (от 1 мин до 1 часа)
	//   Kafka:       MinDelay=1h, MaxDelay=0    (> 1 часа, 0 = бесконечность)
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

type MessageStorage interface {
	Initialize(ctx context.Context, config any) error

	// -- Required methods --
	Store(ctx context.Context, msg *message.Message) error

	FetchReady(ctx context.Context, limit int) ([]*message.Message, error)

	Acknowledge(ctx context.Context, msgID []string) error

	HealthCheck(ctx context.Context) (StorageHealth, error)

	Connect(ctx context.Context) error

	CloseConnection(ctx context.Context) error

	// -- Optional methods --
	GetByID(ctx context.Context, msgID string) (*message.Message, error)
	Count(ctx context.Context) (int64, error)
}
