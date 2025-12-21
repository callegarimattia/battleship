package model

import "github.com/google/uuid"

const gridSize = 10

type cell struct {
	isHit   bool
	shipRef *ship // nil if Water
}

type grid [gridSize][gridSize]cell

type ship struct {
	ShipType
	Hits   int
	IsSunk bool
}

// Player represents a participant in the Battleship game.
// It maintains the state of the player's board (ships) and their tracking board (shots fired).
type Player struct {
	id          string
	board       grid
	tracking    [gridSize][gridSize]ShotResult
	shipsAfloat int
	isReady     bool
	inventory   map[ShipType]int
}

// NewPlayer creates a new Player with a unique ID and a full starting inventory of ships.
// The player starts not ready.
func NewPlayer() *Player {
	return &Player{
		id:        uuid.NewString(),
		inventory: startingInventory(),
		isReady:   false,
	}
}

func (p *Player) String() string {
	return "Player(" + p.id + ")"
}

// ID returns the player's unique identifier.
func (p *Player) ID() string {
	return p.id
}

// PlaceShip allows a player to place a ship on their board.
func (p *Player) PlaceShip(sType ShipType, start Coordinate, o Orientation) error {
	segments := calculateSegments(start, sType.Size(), o)

	if !p.board.inBounds(segments[len(segments)-1]) {
		return ErrOutOfBounds
	}

	if p.board.hasCollision(segments) {
		return ErrShipOverlap
	}

	if p.inventory[sType] <= 0 {
		return ErrShipTypeDepleted
	}

	newShip := &ship{ShipType: sType}
	for _, c := range segments {
		cell := p.board.cellAt(c)
		cell.shipRef = newShip
	}

	p.inventory[sType]--
	p.shipsAfloat++

	return nil
}

// ReceiveAttack handles the incoming shot logic.
func (p *Player) ReceiveAttack(c Coordinate) (ShotResult, error) {
	target := p.board.cellAt(c)
	if target == nil {
		return ResultMiss, ErrOutOfBounds
	}

	if target.isHit && target.shipRef == nil {
		return ResultMiss, nil // Allow repeated misses without error
	}

	if target.isHit && target.shipRef != nil {
		return ResultInvalid, ErrRepeatedHit
	}

	target.isHit = true

	if target.shipRef == nil {
		return ResultMiss, nil
	}

	return p.processShipDamage(target.shipRef)
}

// MarksShotResult updates the player's tracking grid based on the shot result.
func (p *Player) MarksShotResult(c Coordinate, result ShotResult) error {
	target := p.board.cellAt(c)
	if target == nil {
		return ErrOutOfBounds
	}

	switch result {
	case ResultMiss, ResultHit, ResultSunk:
		p.tracking[c.X][c.Y] = result
	default:
		return ErrInvalidShotResult
	}

	return nil
}

// SetReady sets the player's ready status.
// It's idempotent and checks fleet completeness.
func (p *Player) SetReady() error {
	if !p.hasPlacedAllShips() {
		return ErrFleetIncomplete
	}

	p.isReady = true

	return nil
}

// IsReady returns whether the player is ready to start the game.
func (p *Player) IsReady() bool {
	return p.isReady
}

// HasLost returns true if the player has no more ships afloat.
func (p *Player) HasLost() bool {
	return p.shipsAfloat == 0
}

func (p *Player) processShipDamage(s *ship) (ShotResult, error) {
	s.Hits++
	if s.Hits < s.Size() {
		return ResultHit, nil
	}

	s.IsSunk = true
	p.shipsAfloat--

	return ResultSunk, nil
}

func (g *grid) inBounds(c Coordinate) bool {
	return c.X >= 0 && c.X < gridSize && c.Y >= 0 && c.Y < gridSize
}

func (g *grid) cellAt(c Coordinate) *cell {
	if !g.inBounds(c) {
		return nil
	}

	return &g[c.X][c.Y]
}

func (g *grid) hasCollision(coords []Coordinate) bool {
	for _, c := range coords {
		if !g.inBounds(c) || g[c.X][c.Y].shipRef != nil {
			return true
		}
	}

	return false
}

func (p *Player) hasPlacedAllShips() bool {
	for _, count := range p.inventory {
		if count > 0 {
			return false
		}
	}
	return true
}

func startingInventory() map[ShipType]int {
	return map[ShipType]int{
		Carrier:    1,
		Battleship: 1,
		Cruiser:    1,
		Submarine:  1,
		Destroyer:  1,
	}
}
