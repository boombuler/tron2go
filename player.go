package main

import (
	"log"
	"math"
	"sort"
)

type Player struct {
	Id    int
	Score uint
	Name  string
	*playerState
	conn   *connection
	server *GameServer
	input  *InputQueue
}

type playerState struct {
	X         int
	Y         int
	Direction Direction
	Alive     bool
}

type Direction int

const (
	NONE  Direction = 0
	Up    Direction = 1
	Down  Direction = 2
	Left  Direction = 3
	Right Direction = 4
)

func NewPlayer(c *connection, id int, gs *GameServer) *Player {
	result := &Player{
		conn:   c,
		Id:     id,
		server: gs,
		Name:   "",
		Score:  0,
		input:  new(InputQueue),
	}
	go result.readInput()
	return result
}

func (p *Player) Reset(alive bool) {
	playerIds := []int{}

	for _, player := range p.server.Clients {
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
		X:         x,
		Y:         y,
		Direction: NONE,
		Alive:     alive,
	}
	if alive {
		p.server.Board[x][y] = p
	}
	p.playerState = state
}

func (p *Player) AcceptInput() {
	newDir := p.input.Pop()
	switch newDir {
	case Right:
		if p.Direction != Left {
			p.Direction = Right
		}
	case Left:
		if p.Direction != Right {
			p.Direction = Left
		}
	case Up:
		if p.Direction != Down {
			p.Direction = Up
		}
	case Down:
		if p.Direction != Up {
			p.Direction = Down
		}
	}
}

func (p *Player) pushNewDirection(dir Direction) {
	if !p.server.IsRunning {
		p.input.Clear()
	}

	if p.input.count > 0 {
		if p.input.Last() != dir {
			p.input.Push(dir)
		}
	} else {
		p.input.Push(dir)
	}
}

func (p *Player) readInput() {
	for data := range p.conn.receive {
		msgData := newRawJSON(data)

		var cmd string
		if !msgData.getValue("Cmd", &cmd) {
			continue
		}
		log.Println(cmd)
		switch cmd {
		case "move.left":
			p.pushNewDirection(Left)
		case "move.right":
			p.pushNewDirection(Right)
		case "move.up":
			p.pushNewDirection(Up)
		case "move.down":
			p.pushNewDirection(Down)
		case "set.name":
			var name string
			if msgData.getValue("Name", &name) {
				p.Name = name
				p.server.SendInitialState(nil) // Should be replaced with something that sends only the name
			}
		}
	}
}
