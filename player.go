package main

import (
    "math"
    "sort"
    "log"
    "sync"
)

type player struct {
    Id    int
    Score uint
    Name  string
    state *playerState
    conn  *connection
    gs    *gameserver
    input *InputQueue
}

type playerState struct {
    X                 int
    Y                 int
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
        input: &InputQueue{mutex: &sync.Mutex{}, nodes: make([]*InputNode, 10)},
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
        acceptedDirection: NONE,
        alive:             alive,
    }
    if alive {
        p.gs.State.Board[x][y] = p
    }
    p.state = state
}

func (p *player) acceptDirection() {
	newDir := p.input.Pop()
  	switch newDir {
  	case Right:
		if p.state.acceptedDirection != Left {
        	p.state.acceptedDirection = Right
        }
  	case Left:
		if p.state.acceptedDirection != Right {
        	p.state.acceptedDirection = Left
        }
  	case Up:
		if p.state.acceptedDirection != Down {
        	p.state.acceptedDirection = Up
        }
  	case Down:
		if p.state.acceptedDirection != Up {
        	p.state.acceptedDirection = Down
        }
  	}
}

func (p *player) pushNewDirection(dir Direction) {
	if !p.gs.State.IsRunning {
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

func (p *player) readInput() {
    for data := range p.conn.receive {
        msgData := rawJSON(data)

        var cmd string
        if !msgData.GetValue("Cmd", &cmd) {
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
            if msgData.GetValue("Name", &name) {
                p.Name = name
                p.gs.sendInitialState(nil) // Should be replaced with something that sends only the name
            }
        }
    }
}
