package events

// EventBus is the interface for publishing and subscribing to game events.
type EventBus interface {
	// Publish publishes an event to all subscribers.
	Publish(event *GameEvent)
	// Subscribe subscribes to events. Use "*" for matchID to subscribe to all matches.
	Subscribe(matchID string, handler EventHandler) Subscription
	// Close closes the event bus and unsubscribes all subscribers.
	Close()
}

// EventHandler is a function that handles game events.
type EventHandler func(event *GameEvent)

// Subscription represents a subscription to events.
type Subscription interface {
	// Unsubscribe unsubscribes from events.
	Unsubscribe()
}
