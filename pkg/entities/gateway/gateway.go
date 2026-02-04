package gateway

import (
	"context"
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/node"
)

// Info описывает метаданные gateway-узла в системе.
type Info struct {
	ID            string
	Address       string
	Status        node.NodeStatus
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

// Gateway — сущность которая принимает сообщения
// и распределяет их по хранилищам кластера.
type Gateway interface {
	// Consume принимает сообщение и направляет его в соответствующее хранилище
	// на основе ScheduledAt или отправляет в пушеры если сообщение готово.
	Consume(message *message.Message) error
	// Запускает фоновые задачи:
	//
	// - Обновление информации по хранилищам
	Start(ctx context.Context)
	GetConfig() *GatewayConfig
}
