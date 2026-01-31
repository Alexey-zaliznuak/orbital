package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
	routingrule "github.com/Alexey-zaliznuak/orbital/pkg/entities/routing_rule"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrNoSuitableStorage = errors.New("no suitable storage for delay")
)

// Префиксы ключей в etcd
const (
	keyPrefixNodes        = "/orbital/nodes/"
	keyPrefixGateways     = "/orbital/gateways/"
	keyPrefixStorages     = "/orbital/storages/"
	keyPrefixPushers      = "/orbital/pushers/"
	keyPrefixRoutingRules = "/orbital/routing-rules/"
	keyCoordinatorConfig  = "/orbital/coordinators-config"
	keyClusterConfig      = "/orbital/cluster-config"
)

// Storage реализует coordinator.CoordinatorStorage на базе etcd.
type Storage struct {
	client *clientv3.Client
	// timeout для операций с etcd
	timeout time.Duration
}

// Config конфигурация для подключения к etcd.
type Config struct {
	Endpoints   []string
	DialTimeout time.Duration
	OpTimeout   time.Duration
}

// New создаёт новое хранилище координатора на базе etcd.
func New(cfg Config) (*Storage, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	timeout := cfg.OpTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &Storage{
		client:  client,
		timeout: timeout,
	}, nil
}

// Close закрывает соединение с etcd.
func (s *Storage) Close() error {
	return s.client.Close()
}

// === Coordinator Nodes ===

func (s *Storage) CreateNode(ctx context.Context, node *coordinator.Node) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixNodes + node.ID().String()

	data, err := json.Marshal(nodeToDTO(node))
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}

	// Используем транзакцию для проверки существования
	txnResp, err := s.client.Txn(ctx).
		If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, string(data))).
		Commit()
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	if !txnResp.Succeeded {
		return ErrAlreadyExists
	}

	return nil
}

func (s *Storage) GetNode(ctx context.Context, nodeID uuid.UUID) (*coordinator.Node, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixNodes + nodeID.String()

	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}

	var dto nodeDTO
	if err := json.Unmarshal(resp.Kvs[0].Value, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node: %w", err)
	}

	return dtoToNode(&dto), nil
}

func (s *Storage) ListNodes(ctx context.Context) ([]*coordinator.Node, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, keyPrefixNodes, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := make([]*coordinator.Node, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var dto nodeDTO
		if err := json.Unmarshal(kv.Value, &dto); err != nil {
			continue // skip invalid entries
		}
		nodes = append(nodes, dtoToNode(&dto))
	}

	return nodes, nil
}

func (s *Storage) UpdateNodeHeartbeat(ctx context.Context, nodeID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixNodes + nodeID.String()

	// Получаем текущую ноду
	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound
	}

	var dto nodeDTO
	if err := json.Unmarshal(resp.Kvs[0].Value, &dto); err != nil {
		return fmt.Errorf("failed to unmarshal node: %w", err)
	}

	dto.LastHeartbeat = time.Now()
	dto.Status = int(coordinator.NodeStatusActive)

	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}

	_, err = s.client.Put(ctx, key, string(data))
	if err != nil {
		return fmt.Errorf("failed to update node heartbeat: %w", err)
	}

	return nil
}

func (s *Storage) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixNodes + nodeID.String()

	resp, err := s.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	if resp.Deleted == 0 {
		return ErrNotFound
	}

	return nil
}

// === Gateways ===

func (s *Storage) RegisterGateway(ctx context.Context, gateway *coordinator.GatewayInfo) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixGateways + gateway.ID

	data, err := json.Marshal(gateway)
	if err != nil {
		return fmt.Errorf("failed to marshal gateway: %w", err)
	}

	txnResp, err := s.client.Txn(ctx).
		If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, string(data))).
		Commit()
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	if !txnResp.Succeeded {
		return ErrAlreadyExists
	}

	return nil
}

func (s *Storage) GetGateway(ctx context.Context, gatewayID string) (*coordinator.GatewayInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixGateways + gatewayID

	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}

	var gateway coordinator.GatewayInfo
	if err := json.Unmarshal(resp.Kvs[0].Value, &gateway); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gateway: %w", err)
	}

	return &gateway, nil
}

func (s *Storage) ListGateways(ctx context.Context) ([]*coordinator.GatewayInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, keyPrefixGateways, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list gateways: %w", err)
	}

	gateways := make([]*coordinator.GatewayInfo, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var gateway coordinator.GatewayInfo
		if err := json.Unmarshal(kv.Value, &gateway); err != nil {
			continue
		}
		gateways = append(gateways, &gateway)
	}

	return gateways, nil
}

func (s *Storage) UpdateGatewayHeartbeat(ctx context.Context, gatewayID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixGateways + gatewayID

	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get gateway: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound
	}

	var gateway coordinator.GatewayInfo
	if err := json.Unmarshal(resp.Kvs[0].Value, &gateway); err != nil {
		return fmt.Errorf("failed to unmarshal gateway: %w", err)
	}

	gateway.LastHeartbeat = time.Now()
	gateway.Status = coordinator.NodeStatusActive

	data, err := json.Marshal(gateway)
	if err != nil {
		return fmt.Errorf("failed to marshal gateway: %w", err)
	}

	_, err = s.client.Put(ctx, key, string(data))
	return err
}

func (s *Storage) UnregisterGateway(ctx context.Context, gatewayID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixGateways + gatewayID

	resp, err := s.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to unregister gateway: %w", err)
	}

	if resp.Deleted == 0 {
		return ErrNotFound
	}

	return nil
}

// === Storages ===

func (s *Storage) RegisterStorage(ctx context.Context, storage *coordinator.StorageInfo) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixStorages + storage.ID

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal storage: %w", err)
	}

	txnResp, err := s.client.Txn(ctx).
		If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, string(data))).
		Commit()
	if err != nil {
		return fmt.Errorf("failed to register storage: %w", err)
	}

	if !txnResp.Succeeded {
		return ErrAlreadyExists
	}

	return nil
}

func (s *Storage) GetStorage(ctx context.Context, storageID string) (*coordinator.StorageInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixStorages + storageID

	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}

	var storage coordinator.StorageInfo
	if err := json.Unmarshal(resp.Kvs[0].Value, &storage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage: %w", err)
	}

	return &storage, nil
}

func (s *Storage) ListStorages(ctx context.Context) ([]*coordinator.StorageInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, keyPrefixStorages, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list storages: %w", err)
	}

	storages := make([]*coordinator.StorageInfo, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var storage coordinator.StorageInfo
		if err := json.Unmarshal(kv.Value, &storage); err != nil {
			continue
		}
		storages = append(storages, &storage)
	}

	return storages, nil
}

func (s *Storage) UpdateStorageHeartbeat(ctx context.Context, storageID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixStorages + storageID

	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get storage: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound
	}

	var storage coordinator.StorageInfo
	if err := json.Unmarshal(resp.Kvs[0].Value, &storage); err != nil {
		return fmt.Errorf("failed to unmarshal storage: %w", err)
	}

	storage.LastHeartbeat = time.Now()
	storage.Status = coordinator.NodeStatusActive

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal storage: %w", err)
	}

	_, err = s.client.Put(ctx, key, string(data))
	return err
}

func (s *Storage) UnregisterStorage(ctx context.Context, storageID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixStorages + storageID

	resp, err := s.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to unregister storage: %w", err)
	}

	if resp.Deleted == 0 {
		return ErrNotFound
	}

	return nil
}

// === Pushers ===

func (s *Storage) RegisterPusher(ctx context.Context, pusher *coordinator.PusherInfo) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixPushers + pusher.ID

	data, err := json.Marshal(pusher)
	if err != nil {
		return fmt.Errorf("failed to marshal pusher: %w", err)
	}

	txnResp, err := s.client.Txn(ctx).
		If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, string(data))).
		Commit()
	if err != nil {
		return fmt.Errorf("failed to register pusher: %w", err)
	}

	if !txnResp.Succeeded {
		return ErrAlreadyExists
	}

	return nil
}

func (s *Storage) GetPusher(ctx context.Context, pusherID string) (*coordinator.PusherInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixPushers + pusherID

	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get pusher: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}

	var pusher coordinator.PusherInfo
	if err := json.Unmarshal(resp.Kvs[0].Value, &pusher); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pusher: %w", err)
	}

	return &pusher, nil
}

func (s *Storage) ListPushers(ctx context.Context) ([]*coordinator.PusherInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, keyPrefixPushers, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list pushers: %w", err)
	}

	pushers := make([]*coordinator.PusherInfo, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var pusher coordinator.PusherInfo
		if err := json.Unmarshal(kv.Value, &pusher); err != nil {
			continue
		}
		pushers = append(pushers, &pusher)
	}

	return pushers, nil
}

func (s *Storage) UpdatePusherHeartbeat(ctx context.Context, pusherID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixPushers + pusherID

	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get pusher: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound
	}

	var pusher coordinator.PusherInfo
	if err := json.Unmarshal(resp.Kvs[0].Value, &pusher); err != nil {
		return fmt.Errorf("failed to unmarshal pusher: %w", err)
	}

	pusher.LastHeartbeat = time.Now()
	pusher.Status = coordinator.NodeStatusActive

	data, err := json.Marshal(pusher)
	if err != nil {
		return fmt.Errorf("failed to marshal pusher: %w", err)
	}

	_, err = s.client.Put(ctx, key, string(data))
	return err
}

func (s *Storage) UnregisterPusher(ctx context.Context, pusherID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixPushers + pusherID

	resp, err := s.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to unregister pusher: %w", err)
	}

	if resp.Deleted == 0 {
		return ErrNotFound
	}

	return nil
}

// === Routing Rules ===

func (s *Storage) CreateRoutingRule(ctx context.Context, rule *routingrule.RoutingRule) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixRoutingRules + rule.ID

	data, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("failed to marshal routing rule: %w", err)
	}

	txnResp, err := s.client.Txn(ctx).
		If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, string(data))).
		Commit()
	if err != nil {
		return fmt.Errorf("failed to create routing rule: %w", err)
	}

	if !txnResp.Succeeded {
		return ErrAlreadyExists
	}

	return nil
}

func (s *Storage) GetRoutingRule(ctx context.Context, ruleID string) (*routingrule.RoutingRule, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixRoutingRules + ruleID

	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get routing rule: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}

	var rule routingrule.RoutingRule
	if err := json.Unmarshal(resp.Kvs[0].Value, &rule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routing rule: %w", err)
	}

	return &rule, nil
}

func (s *Storage) ListRoutingRules(ctx context.Context) ([]*routingrule.RoutingRule, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, keyPrefixRoutingRules, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list routing rules: %w", err)
	}

	rules := make([]*routingrule.RoutingRule, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var rule routingrule.RoutingRule
		if err := json.Unmarshal(kv.Value, &rule); err != nil {
			continue
		}
		rules = append(rules, &rule)
	}

	return rules, nil
}

func (s *Storage) UpdateRoutingRule(ctx context.Context, rule *routingrule.RoutingRule) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixRoutingRules + rule.ID

	// Проверяем существование
	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get routing rule: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound
	}

	data, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("failed to marshal routing rule: %w", err)
	}

	_, err = s.client.Put(ctx, key, string(data))
	return err
}

func (s *Storage) DeleteRoutingRule(ctx context.Context, ruleID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	key := keyPrefixRoutingRules + ruleID

	resp, err := s.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete routing rule: %w", err)
	}

	if resp.Deleted == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Storage) GetCoordinatorConfig(ctx context.Context) (*coordinator.CoordinatorConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, keyCoordinatorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get coordinator config: %w", err)
	}

	if len(resp.Kvs) == 0 {
		// Возвращаем дефолтный конфиг
		return &coordinator.CoordinatorConfig{}, nil
	}

	var config coordinator.CoordinatorConfig
	if err := json.Unmarshal(resp.Kvs[0].Value, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal coordinator config: %w", err)
	}

	return &config, nil
}

func (s *Storage) SetCoordinatorConfig(ctx context.Context, config *coordinator.CoordinatorConfig) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal coordinator config: %w", err)
	}

	_, err = s.client.Put(ctx, keyCoordinatorConfig, string(data))
	return err
}

func (s *Storage) GetClusterConfig(ctx context.Context) (*coordinator.ClusterConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, keyClusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config: %w", err)
	}

	if len(resp.Kvs) == 0 {
		// Возвращаем дефолтный конфиг
		return &coordinator.ClusterConfig{}, nil
	}

	var config coordinator.ClusterConfig
	if err := json.Unmarshal(resp.Kvs[0].Value, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cluster config: %w", err)
	}

	return &config, nil
}

func (s *Storage) SetClusterConfig(ctx context.Context, config *coordinator.ClusterConfig) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = s.client.Put(ctx, keyClusterConfig, string(data))
	return err
}
