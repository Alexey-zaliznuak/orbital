package inmemory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
)

var (
	ErrNotFound       = errors.New("message not found")
	ErrNotInitialized = errors.New("storage not initialized")
	ErrAlreadyExists  = errors.New("message with this ID already exists")
)

// InMemoryStorage реализует storage.MessageStorage в оперативной памяти.
// Взятые сообщения помечаются как in-flight и не возвращаются повторно до Acknowledge.
type InMemoryStorage struct {
	mu       sync.RWMutex
	messages map[string]*message.Message
	inflight map[string]struct{}

	config *storage.BaseStorageConfig
	ready  bool
}

// NewInMemoryStorage создаёт незаполненное хранилище.
// Перед использованием необходимо вызвать Initialize.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

// Initialize настраивает хранилище. config должен быть *storage.BaseStorageConfig.
func (s *InMemoryStorage) Initialize(_ context.Context, rawConfig any) error {
	cfg, ok := rawConfig.(*storage.BaseStorageConfig)
	if !ok {
		return fmt.Errorf("expected *storage.BaseStorageConfig, got %T", rawConfig)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = cfg
	s.messages = make(map[string]*message.Message)
	s.inflight = make(map[string]struct{})
	s.ready = true

	return nil
}

// Connect — no-op для in-memory хранилища.
func (s *InMemoryStorage) Connect(_ context.Context) error {
	return nil
}

// CloseConnection — no-op для in-memory хранилища.
func (s *InMemoryStorage) CloseConnection(_ context.Context) error {
	return nil
}

// HealthCheck возвращает StorageHealthOK если хранилище инициализировано.
func (s *InMemoryStorage) HealthCheck(_ context.Context) (storage.StorageHealth, error) {
	return storage.StorageHealthOK, nil
}

// Store сохраняет сообщение. Если ID не задан — генерируется UUIDv6.
// Дублирование по ID не допускается.
func (s *InMemoryStorage) Store(_ context.Context, msg *message.Message) error {
	if err := s.checkReady(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	copied := *msg
	if copied.ID == "" {
		copied.ID = message.GenerateID()
	}

	if _, exists := s.messages[copied.ID]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, copied.ID)
	}

	s.messages[copied.ID] = &copied

	return nil
}

// FetchReady возвращает до limit сообщений, готовых к доставке (ScheduledAt <= now),
// и помечает их как in-flight — повторный FetchReady их не вернёт до Acknowledge.
func (s *InMemoryStorage) FetchReady(_ context.Context, limit int) ([]*message.Message, error) {
	if err := s.checkReady(); err != nil {
		return nil, err
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]*message.Message, 0, limit)

	for id, msg := range s.messages {
		if len(result) >= limit {
			break
		}
		if _, inFlight := s.inflight[id]; inFlight {
			continue
		}
		if !isReady(msg, now) {
			continue
		}
		s.inflight[id] = struct{}{}
		copied := *msg
		result = append(result, &copied)
	}

	return result, nil
}

// Acknowledge подтверждает обработку сообщений и удаляет их из хранилища.
// Неизвестные ID игнорируются.
func (s *InMemoryStorage) Acknowledge(_ context.Context, ids []string) error {
	if err := s.checkReady(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		delete(s.messages, id)
		delete(s.inflight, id)
	}

	return nil
}

// GetByID возвращает сообщение по ID (включая in-flight).
func (s *InMemoryStorage) GetByID(_ context.Context, id string) (*message.Message, error) {
	if err := s.checkReady(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	msg, ok := s.messages[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	copied := *msg
	return &copied, nil
}

// Count возвращает общее количество сообщений (включая in-flight).
func (s *InMemoryStorage) Count(_ context.Context) (int64, error) {
	if err := s.checkReady(); err != nil {
		return 0, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.messages)), nil
}

// isReady — сообщение готово к доставке если ScheduledAt не задан или уже наступил.
func isReady(msg *message.Message, now time.Time) bool {
	return msg.ScheduledAt.IsZero() || !msg.ScheduledAt.After(now)
}

func (s *InMemoryStorage) checkReady() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.ready {
		return ErrNotInitialized
	}
	return nil
}
