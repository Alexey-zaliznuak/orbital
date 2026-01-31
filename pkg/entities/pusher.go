package entities

// Pusher отправляет сообщения во внешнюю систему.
type Pusher interface {
	// Push отправляет сообщение.
	Push(msg *Message) error
}
