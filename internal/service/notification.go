package service

import (
	"sync"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/google/uuid"
)

// NotificationService implements controller.NotificationService
type NotificationService struct {
	subscribers map[string][]subscriber
	mu          sync.RWMutex
}

type subscriber struct {
	id string
	ch chan *dto.GameEvent
}

type subscription struct {
	ns      *NotificationService
	matchID string
	id      string
}

// NewNotificationService creates a new notification service.
func NewNotificationService() *NotificationService {
	return &NotificationService{
		subscribers: make(map[string][]subscriber),
	}
}

// Subscribe returns a channel of events for the match.
func (s *NotificationService) Subscribe(
	matchID string,
) (sub controller.Subscription, out <-chan *dto.GameEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	ch := make(chan *dto.GameEvent, 100)

	s.subscribers[matchID] = append(s.subscribers[matchID],
		subscriber{
			id: id,
			ch: ch,
		})

	return &subscription{
		ns:      s,
		matchID: matchID,
		id:      id,
	}, ch
}

// Publish publishes an event to all subscribers.
func (s *NotificationService) Publish(event *dto.GameEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Notify match-specific subscribers
	s.publishToSlice(event, s.subscribers[event.MatchID])

	// Notify wildcard subscribers (if any, represented by "*")
	s.publishToSlice(event, s.subscribers["*"])
}

func (s *NotificationService) publishToSlice(event *dto.GameEvent, subscribers []subscriber) {
	for _, sub := range subscribers {
		select {
		case sub.ch <- event:
		default:
			// Non-blocking send
		}
	}
}

// Unsubscribe removes the subscription.
func (s *subscription) Unsubscribe() {
	s.ns.mu.Lock()
	defer s.ns.mu.Unlock()

	subs := s.ns.subscribers[s.matchID]
	for i, sub := range subs {
		if sub.id == s.id {
			// Close the channel to signal end of stream
			close(sub.ch)
			s.ns.subscribers[s.matchID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}
