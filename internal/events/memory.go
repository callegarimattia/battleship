package events

import (
	"sync"

	"github.com/google/uuid"
)

// MemoryEventBus is an in-memory implementation of EventBus.
type MemoryEventBus struct {
	subscribers map[string][]subscriber
	mu          sync.RWMutex
	closed      bool
}

type subscriber struct {
	id      string
	handler EventHandler
}

type subscription struct {
	bus     *MemoryEventBus
	matchID string
	id      string
}

// NewMemoryEventBus creates a new in-memory event bus.
func NewMemoryEventBus() *MemoryEventBus {
	return &MemoryEventBus{
		subscribers: make(map[string][]subscriber),
	}
}

// Publish publishes an event to all subscribers.
func (b *MemoryEventBus) Publish(event *GameEvent) { //nolint
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	// Publish to match-specific subscribers
	for _, sub := range b.subscribers[event.MatchID] {
		go sub.handler(event)
	}

	// Publish to wildcard subscribers
	for _, sub := range b.subscribers["*"] {
		go sub.handler(event)
	}
}

// Subscribe subscribes to events for a specific match or all matches ("*").
func (b *MemoryEventBus) Subscribe(matchID string, handler EventHandler) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := uuid.NewString()
	sub := subscriber{
		id:      id,
		handler: handler,
	}

	b.subscribers[matchID] = append(b.subscribers[matchID], sub)

	return &subscription{
		bus:     b,
		matchID: matchID,
		id:      id,
	}
}

// Close closes the event bus and removes all subscribers.
func (b *MemoryEventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	b.subscribers = make(map[string][]subscriber)
}

// Unsubscribe removes the subscription.
func (s *subscription) Unsubscribe() {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()

	subs := s.bus.subscribers[s.matchID]
	for i, sub := range subs {
		if sub.id == s.id {
			s.bus.subscribers[s.matchID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}
