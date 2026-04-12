package inmemory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/bus"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
	"github.com/Alexey-zaliznuak/orbital/pkg/logger"
	natsclient "github.com/Alexey-zaliznuak/orbital/pkg/nats"
	"github.com/Alexey-zaliznuak/orbital/pkg/sdk/coordinator"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var (
	ErrNotFound       = errors.New("message not found")
	ErrNotInitialized = errors.New("storage not initialized")
	ErrAlreadyExists  = errors.New("message with this ID already exists")
)

type dumpData struct {
	Messages map[string]*message.Message `json:"messages"`
	Inflight []string                    `json:"inflight"`
}

type InMemoryStorage struct {
	mu sync.RWMutex

	initOnce     sync.Once
	loadDumpOnce sync.Once

	messagesMu sync.RWMutex
	messages   map[string]*message.Message

	inflightMu sync.RWMutex
	inflight   map[string]*message.Message

	inflightProcessMu    sync.Mutex
	sendExpiredProcessMu sync.Mutex

	busClient *bus.Client

	cfg   *InMemoryStorageConfig
	ready bool
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

func (s *InMemoryStorage) Initialize(ctx context.Context, rawConfig any) error {
	var err error

	s.initOnce.Do(func() {
		err = s.initialize(ctx, rawConfig)
		if err != nil {
			logger.Log.Error("failed to initialize storage", zap.Error(err))
		}
	})

	return err
}

func (s *InMemoryStorage) initialize(ctx context.Context, rawConfig any) error {
	if s.ready {
		return nil
	}

	cfg, ok := rawConfig.(*InMemoryStorageConfig)

	if !ok {
		return fmt.Errorf("expected *InMemoryStorageConfig, got %T", rawConfig)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	coordinatorClient := coordinator.NewClient(coordinator.ClientConfig{
		BaseURL: cfg.ClusterAddress,
	})

	clusterCfg, err := coordinatorClient.GetClusterConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
	}

	natsClient, err := natsclient.New(natsclient.Config{
		URL: clusterCfg.NatsAddress,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	s.busClient = bus.New(natsClient)

	s.cfg = cfg
	s.messages = make(map[string]*message.Message)
	s.inflight = make(map[string]*message.Message)
	s.ready = true

	s.loadDumpOnce.Do(func() {
		if !cfg.UseDump {
			return
		}
		if err := s.LoadFromDump(ctx); err != nil {
			logger.Log.Error("failed to load dump", zap.Error(err))
		}
	})

	findExpiredTicker := time.NewTicker(cfg.FindExpiredInterval)
	sendExpiredTicker := time.NewTicker(cfg.SendExpiredInterval)

	s.busClient.NewHandlerOnStorageMessages(cfg.ID, s.HandleNewMessages)

	go func() {
		defer findExpiredTicker.Stop()
		defer sendExpiredTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				if s.cfg.UseDump {
					if err := s.Dump(ctx); err != nil {
						logger.Log.Error("failed to dump", zap.Error(err))
					}
				}
				return
			case <-findExpiredTicker.C:
				go func() {
					if ok := s.inflightProcessMu.TryLock(); !ok {
						return
					}
					defer s.inflightProcessMu.Unlock()

					if err := s.moveExpiredToInflight(ctx); err != nil {
						logger.Log.Error("failed to move ready to inflight", zap.Error(err))
					}
				}()
			case <-sendExpiredTicker.C:
				go func() {
					if ok := s.sendExpiredProcessMu.TryLock(); !ok {
						return
					}
					defer s.sendExpiredProcessMu.Unlock()

					if err := s.processMessages(ctx); err != nil {
						logger.Log.Error("failed to process expired messages", zap.Error(err))
					}
				}()
			}
		}
	}()

	return nil
}

func (s *InMemoryStorage) HealthCheck(_ context.Context) (storage.StorageHealth, error) {
	if err := s.checkReady(); err != nil {
		return storage.StorageHealthDisconnect, err
	}
	return storage.StorageHealthOK, nil
}

func (s *InMemoryStorage) Store(_ context.Context, msgs []*message.Message) error {
	if err := s.checkReady(); err != nil {
		return err
	}

	s.messagesMu.Lock()
	defer s.messagesMu.Unlock()

	for _, msg := range msgs {
		copied := *msg
		if copied.ID == "" {
			copied.ID = message.GenerateID()
		}

		if _, exists := s.messages[copied.ID]; exists {
			return fmt.Errorf("%w: %s", ErrAlreadyExists, copied.ID)
		}

		s.messages[copied.ID] = &copied
	}

	return nil
}

func (s *InMemoryStorage) HandleNewMessages(msg *nats.Msg) {
	msgs := make([]*message.Message, 0)

	if err := json.Unmarshal(msg.Data, &msgs); err != nil {
		logger.Log.Error("Received messages unmarshal error", zap.Error(err))
		return
	}

	if err := s.Store(context.Background(), msgs); err != nil {
		logger.Log.Error("Failed to store messages", zap.Error(err))
	}
}

func (s *InMemoryStorage) processMessages(ctx context.Context) error {
	if err := s.checkReady(); err != nil {
		return err
	}

	s.inflightMu.RLock()

	batchSize := int(math.Min(float64(len(s.inflight)), float64(s.cfg.MaxOutputBatchSize)))

	if batchSize == 0 {
		return nil
	}

	ids := make([]string, 0, batchSize)
	msgs := make([]*message.Message, 0, batchSize)

	for _, msg := range s.inflight {
		msgs = append(msgs, msg)
		ids = append(ids, msg.ID)

		if len(msgs) >= batchSize {
			break
		}
	}

	s.inflightMu.RUnlock()

	if err := s.busClient.SendToGateway(msgs); err != nil {
		return err
	}

	if err := s.acknowledge(ctx, ids); err != nil {
		return err
	}

	return nil
}

func (s *InMemoryStorage) moveExpiredToInflight(_ context.Context) error {
	if err := s.checkReady(); err != nil {
		return err
	}

	now := time.Now()

	s.messagesMu.RLock()
	defer s.messagesMu.RUnlock()

	s.inflightMu.Lock()
	defer s.inflightMu.Unlock()

	for id, msg := range s.messages {
		if _, inFlight := s.inflight[id]; inFlight {
			continue
		}
		if !isReady(msg, now) {
			continue
		}
		s.inflight[id] = msg
	}

	return nil
}

func (s *InMemoryStorage) acknowledge(_ context.Context, ids []string) error {
	if err := s.checkReady(); err != nil {
		return err
	}

	s.messagesMu.Lock()
	defer s.messagesMu.Unlock()

	s.inflightMu.Lock()
	defer s.inflightMu.Unlock()

	for _, id := range ids {
		delete(s.inflight, id)
		delete(s.messages, id)
	}

	return nil
}

func (s *InMemoryStorage) GetByID(_ context.Context, id string) (*message.Message, error) {
	if err := s.checkReady(); err != nil {
		return nil, err
	}

	s.messagesMu.RLock()
	defer s.messagesMu.RUnlock()

	msg, ok := s.messages[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	copied := *msg
	return &copied, nil
}

func (s *InMemoryStorage) Count(_ context.Context) (int64, error) {
	if err := s.checkReady(); err != nil {
		return 0, err
	}

	s.messagesMu.RLock()
	defer s.messagesMu.RUnlock()

	return int64(len(s.messages)), nil
}

func (s *InMemoryStorage) Dump(_ context.Context) error {
	if err := s.checkReady(); err != nil {
		return err
	}

	s.messagesMu.RLock()
	defer s.messagesMu.RUnlock()

	s.inflightMu.RLock()
	defer s.inflightMu.RUnlock()

	inflight := make([]string, 0, len(s.inflight))

	for id := range s.inflight {
		inflight = append(inflight, id)
	}

	data := dumpData{
		Messages: s.messages,
		Inflight: inflight,
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("dump marshal: %w", err)
	}

	if err := os.WriteFile(s.cfg.DumpFile, raw, 0o644); err != nil {
		return fmt.Errorf("dump write %s: %w", s.cfg.DumpFile, err)
	}

	return nil
}

func (s *InMemoryStorage) LoadFromDump(_ context.Context) error {
	raw, err := os.ReadFile(s.cfg.DumpFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("dump read %s: %w", s.cfg.DumpFile, err)
	}

	var data dumpData
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("dump unmarshal: %w", err)
	}

	s.messagesMu.Lock()
	defer s.messagesMu.Unlock()

	s.inflightMu.Lock()
	defer s.inflightMu.Unlock()

	s.messages = data.Messages
	if s.messages == nil {
		s.messages = make(map[string]*message.Message)
	}

	s.inflight = make(map[string]*message.Message, len(data.Inflight))

	for _, id := range data.Inflight {
		msg, ok := data.Messages[id]
		if !ok {
			continue
		}
		s.inflight[id] = msg
	}

	return nil
}

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
