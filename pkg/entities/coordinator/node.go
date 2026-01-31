package coordinator

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type NodeStatus int

const (
	NodeStatusConnecting NodeStatus = iota
	NodeStatusActive
	NodeStatusRemoved
)

func (status NodeStatus) String() string {
	switch status {
	case NodeStatusConnecting:
		return "Connecting"
	case NodeStatusActive:
		return "Active"
	case NodeStatusRemoved:
		return "Removed"
	default:
		return "Unknown"
	}
}

// Node представляет узел в распределённой системе
type Node struct {
	mu sync.RWMutex

	id      uuid.UUID
	address string
	status  NodeStatus

	registeredAt  time.Time
	lastHeartbeat time.Time
}

// NewNode создаёт новую ноду с заданным ID.
func NewNode(id uuid.UUID, address string) *Node {
	now := time.Now()
	return &Node{
		id:            id,
		address:       address,
		status:        NodeStatusConnecting,
		registeredAt:  now,
		lastHeartbeat: now,
	}
}

// NewNodeFromDTO восстанавливает ноду из хранилища.
func NewNodeFromDTO(
	id uuid.UUID,
	address string,
	status NodeStatus,
	registeredAt time.Time,
	lastHeartbeat time.Time,
) *Node {
	return &Node{
		id:            id,
		address:       address,
		status:        status,
		registeredAt:  registeredAt,
		lastHeartbeat: lastHeartbeat,
	}
}

// ID возвращает идентификатор ноды
func (n *Node) ID() uuid.UUID {
	return n.id
}

// Address возвращает адрес ноды
func (n *Node) Address() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.address
}

// Status возвращает текущий статус ноды
func (n *Node) Status() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status
}


// RegisteredAt возвращает время регистрации ноды
func (n *Node) RegisteredAt() time.Time {
	return n.registeredAt
}

// LastHeartbeat возвращает время последнего heartbeat
func (n *Node) LastHeartbeat() time.Time {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.lastHeartbeat
}

// IsAlive проверяет, жива ли нода (heartbeat в пределах timeout)
func (n *Node) IsAlive(timeout time.Duration) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status == NodeStatusActive && time.Since(n.lastHeartbeat) < timeout
}

// IsActive проверяет, активна ли нода
func (n *Node) IsActive() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status == NodeStatusActive
}
