package entities

// Gateway — сущность которая принимает сообщения
// и распределяет их по хранилищам кластера.
type Gateway interface {
	// Consume принимает сообщение и направляет его в соответствующее хранилище
	// на основе ScheduledAt или отправляет в пушеры если сообщение готово.
	Consume(message *Message) error
}
