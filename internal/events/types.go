package events

import "time"

// EventType represents the type of game event.
type EventType string

// EvenTpe possible values
const (
	EventPlayerJoined EventType = "player.joined"
	EventShipPlaced   EventType = "ship.placed"
	EventAttackMade   EventType = "attack.made"
	EventGameStarted  EventType = "game.started"
	EventGameOver     EventType = "game.over"
	EventTurnChanged  EventType = "turn.changed"
)

// GameEvent represents a game event that can be published to subscribers.
type GameEvent struct {
	Type      EventType
	MatchID   string
	PlayerID  string // Player who triggered the event
	TargetID  string // Player who should be notified
	Data      any
	Timestamp time.Time
}

// AttackEventData contains data for attack events.
type AttackEventData struct {
	X      int
	Y      int
	Result string // "hit", "miss", "sunk"
}

// ShipPlacedEventData contains data for ship placement events.
type ShipPlacedEventData struct {
	Size     int
	X        int
	Y        int
	Vertical bool
}

// GameOverEventData contains data for game over events.
type GameOverEventData struct {
	Winner string
}
