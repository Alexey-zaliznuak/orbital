package entities

// EventType определяет тип события.
type EventType int

const (
	// EventCreated — объект создан.
	EventCreated EventType = iota
	// EventUpdated — объект обновлён.
	EventUpdated
	// EventDeleted — объект удалён.
	EventDeleted
)
