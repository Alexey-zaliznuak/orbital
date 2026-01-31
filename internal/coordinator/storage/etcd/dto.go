package etcd

import (
	"time"

	"github.com/google/uuid"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
)

// nodeDTO — DTO для сериализации Node в etcd.
// Необходим, так как Node имеет приватные поля.
type nodeDTO struct {
	ID            uuid.UUID `json:"id"`
	Address       string    `json:"address"`
	Status        int       `json:"status"`
	RegisteredAt  time.Time `json:"registered_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

func nodeToDTO(n *coordinator.Node) *nodeDTO {
	return &nodeDTO{
		ID:            n.ID(),
		Address:       n.Address(),
		Status:        int(n.Status()),
		RegisteredAt:  n.RegisteredAt(),
		LastHeartbeat: n.LastHeartbeat(),
	}
}

func dtoToNode(dto *nodeDTO) *coordinator.Node {
	return coordinator.NewNodeFromDTO(
		dto.ID,
		dto.Address,
		coordinator.NodeStatus(dto.Status),
		dto.RegisteredAt,
		dto.LastHeartbeat,
	)
}
