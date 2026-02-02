package entities

import "github.com/Alexey-zaliznuak/orbital/pkg/entities/message"

// Pusher отправляет сообщения во внешнюю систему.
type Pusher interface {
	// Push отправляет сообщение.
	Push(msg *message.Message) error
}
