package main

import (
	"log"
	"math"
	"sort"
)

type Client struct {
	Id    int
	Score uint
	Name  string
	kind  ClientKind
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

type Direction byte

const (
	NONE Direction = iota
	Up
	Down
	Left
	Right
)

type ClientKind byte

const (
	Spectator ClientKind = iota
	Player
)

func (ct ClientKind) String() string {
	switch ct {
	case Spectator:
		return "Spectator"
	case Player:
		return "Player"
	}
	log.Fatal("Invalid ClientKind: ", ct)
	return ""
}

func NewClient(c *connection, gs *GameServer) *Client {
	result := &Client{
		conn:   c,
		Id:     -1,
		server: gs,
		kind:   Spectator,
		Name:   "",
		Score:  0,
		input:  new(InputQueue),
	}
	go result.readInput()
	return result
}

func (c *Client) Reset(alive bool) {
	if c.kind != Player {
		return
	}

	playerIds := []int{}

	for _, client := range c.server.Clients {
		if client.kind == Player {
			playerIds = append(playerIds, client.Id)
		}
	}

	sort.Ints(playerIds)
	idx := sort.SearchInts(playerIds, c.Id)

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
		c.server.Board[x][y] = c
	}
	c.playerState = state
}

func (c *Client) AcceptInput() {
	newDir := c.input.Pop()
	switch newDir {
	case Right:
		if c.Direction != Left {
			c.Direction = Right
		}
	case Left:
		if c.Direction != Right {
			c.Direction = Left
		}
	case Up:
		if c.Direction != Down {
			c.Direction = Up
		}
	case Down:
		if c.Direction != Up {
			c.Direction = Down
		}
	}
}

func (c *Client) pushNewDirection(dir Direction) {
	if !c.server.IsRunning {
		c.input.Clear()
	}

	if c.input.count > 0 {
		if c.input.Last() != dir {
			c.input.Push(dir)
		}
	} else {
		c.input.Push(dir)
	}
}

func (c *Client) tryElevateTo(kind ClientKind) {
	switch kind {
	case Spectator:
		if c.kind == Player {
			c.kind = Spectator
			c.server.idStore.free <- c.Id
			c.Id = -1
			if c.Alive {
				c.Alive = false
				if !c.server.IsRunning {
					c.server.newGame()
				}
			}
		}
	case Player:
		c.Id = c.server.idStore.TryGet()
		if c.Id >= 0 {
			c.kind = Player
			if c.server.IsRunning {
				c.Reset(false)
			} else {
				c.server.newGame()
			}
		}
	}
	c.conn.send <- c.SerializeIdentity()
}

func (c *Client) readInput() {
	for data := range c.conn.receive {
		msgData := newRawJSON(data)

		var cmd string
		if !msgData.getValue("Cmd", &cmd) {
			continue
		}
		log.Println(cmd)
		switch cmd {
		case "move.left":
			c.pushNewDirection(Left)
		case "move.right":
			c.pushNewDirection(Right)
		case "move.up":
			c.pushNewDirection(Up)
		case "move.down":
			c.pushNewDirection(Down)
		case "set.name":
			var name string
			if msgData.getValue("Name", &name) {
				c.Name = name
				if name == "" {
					c.tryElevateTo(Spectator)
				} else {
					c.tryElevateTo(Player)
				}
				c.server.SendInitialState(nil) // Should be replaced with something that sends only the name
			}
		}
	}
}
