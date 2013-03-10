package main

import (
	"math"
	"sort"
)

type player struct {
	Id    int
	Score uint
	Name  string
	state *playerState
	conn  *connection
	gs    *gameserver
}

type playerState struct {
	X                 int
	Y                 int
	direction         Direction
	acceptedDirection Direction
	alive             bool
}

type Direction int

const (
	NONE  Direction = 0
	Up    Direction = 1
	Down  Direction = 2
	Left  Direction = 3
	Right Direction = 4
)

func createPlayer(c *connection, id int, gs *gameserver) *player {
	result := &player{
		conn:  c,
		Id:    id,
		gs:    gs,
		Name:  "",
		Score: 0,
	}
	go result.readInput()
	return result
}

func (p *player) newPlayerState(alive bool) {
	playerIds := []int{}

	for _, player := range p.gs.clients {
		playerIds = append(playerIds, player.Id)
	}

	sort.Ints(playerIds)
	idx := sort.SearchInts(playerIds, p.Id)

	centerX := float64(FIELD_WIDTH) / 2.0
	centerY := float64(FIELD_HEIGHT) / 2.0

	r := (math.Min(float64(FIELD_WIDTH), float64(FIELD_HEIGHT)) / 2.0) * 0.7
	deg := (math.Pi / float64(len(playerIds))) * (2.0 * float64(idx))

	x := int(math.Ceil(centerX + r*math.Cos(deg)))
	y := int(math.Ceil(centerY + r*math.Sin(deg)))

	state := &playerState{
		X:                 x,
		Y:                 y,
		direction:         NONE,
		acceptedDirection: NONE,
		alive:             alive,
	}
	p.gs.State.Board[x][y] = p
	p.state = state
}

func (p *player) acceptDirection() {
	p.state.acceptedDirection = p.state.direction
}

func (p *player) readInput() {
	for msg := range p.conn.receive {
		switch string(msg) {
		case "move.left":
			if p.state.acceptedDirection != Right {
				p.state.direction = Left
			}
		case "move.right":
			if p.state.acceptedDirection != Left {
				p.state.direction = Right
			}
		case "move.up":
			if p.state.acceptedDirection != Down {
				p.state.direction = Up
			}
		case "move.down":
			if p.state.acceptedDirection != Up {
				p.state.direction = Down
			}
		}
	}
}
