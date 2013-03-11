package main

import (
	"log"
	"time"
)

type GameServer struct {
	Clients    map[*connection]*Client
	Register   chan *connection
	Unregister chan *connection
	Broadcast  chan []byte
	idStore    *idStore
	*gameState
}

var gameserver *GameServer = NewGameServer()

func NewGameServer() *GameServer {
	result := &GameServer{
		Register:   make(chan *connection),
		Unregister: make(chan *connection),
		Clients:    make(map[*connection]*Client),
		Broadcast:  make(chan []byte),
		idStore:    createIdStore(),
	}
	result.newGame()
	return result
}

func (gs *GameServer) newGame() {
	log.Println("NewGame")
	gs.gameState = NewGameState()
	for _, p := range gs.Clients {
		p.Reset(true)
	}
	gs.SendInitialState(nil)
}

func (gs *GameServer) gameLoop(endSignal chan bool) {
	ticker := time.NewTicker(SPEED)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			gs.calcRound()

		case <-endSignal:
			return
		}
	}
}

func (gs *GameServer) SendInitialState(c *connection) {
	data := &GameStateData{Blocks: make([]NewBlock, 0), Players: make([]Client, 0)}

	for _, p := range gs.Clients {
		if p.kind == Player {
			data.Players = append(data.Players, *p)
		}
	}

	for x, col := range gs.Board {
		for y, p := range col {
			if p != nil {
				data.Blocks = append(data.Blocks, *&NewBlock{X: x, Y: y, PlayerId: p.Id})
			}
		}
	}
	if c != nil {
		c.send <- data.Serialize()
	} else if len(gs.Clients) > 0 {
		gs.Broadcast <- data.Serialize()
	}
}

func (gs *GameServer) getPlayers() []*Client {
	res := make([]*Client, 0)
	for _, c := range gs.Clients {
		if c.kind == Player {
			res = append(res, c)
		}
	}
	return res
}

func (gs *GameServer) calcRound() {
	players := gs.getPlayers()

	if !gs.IsRunning {
		if len(players) == 0 {
			return
		}
		for _, p := range players {
			p.AcceptInput()
			if p.Direction == NONE {
				return
			}
		}
		gs.IsRunning = true
		return
	}

	roundData := &RoundData{Blocks: make([]NewBlock, 0)}

	for _, p := range players {
		if p.Alive {
			blk := gs.movePlayer(p)
			if blk != nil {
				roundData.Blocks = append(roundData.Blocks, *blk)
			}

		}
	}
	gs.Broadcast <- roundData.Serialize()

	gs.checkGameOver()
}

func (gs *GameServer) checkGameOver() {
	alivecount := 0
	for _, p := range gs.Clients {
		if p.kind == Player && p.Alive {
			alivecount++
		}
	}

	// if alivecount < 2 {
	//     gs.newGame()
	// }
}

func (gs *GameServer) movePlayer(p *Client) *NewBlock {
	if p.kind != Player {
		return nil
	}
	p.AcceptInput()
	switch p.Direction {
	case NONE:
		return nil
	case Left:
		p.X = p.X - 1
	case Right:
		p.X = p.X + 1
	case Up:
		p.Y = p.Y - 1
	case Down:
		p.Y = p.Y + 1
	}
	if p.X < 0 || p.Y < 0 || p.X >= FIELD_WIDTH || p.Y >= FIELD_HEIGHT ||
		gs.Board[p.X][p.Y] != nil {
		p.Alive = false
		return nil
	} else {
		gs.Board[p.X][p.Y] = p
	}

	return &NewBlock{X: p.X, Y: p.Y, PlayerId: p.Id}
}

func (gs *GameServer) run() {
	go gs.gameLoop(nil)
	for {
		select {
		case c := <-gs.Register:
			go gs.onPlayerConnected(c)
		case c := <-gs.Unregister:
			if gs.Clients[c].Id > 0 {
				gs.idStore.free <- gs.Clients[c].Id
			}
			delete(gs.Clients, c)
			close(c.send)
			close(c.receive)
		case m := <-gs.Broadcast:
			for c, _ := range gs.Clients {
				c.send <- m
			}
		}
	}
}

func (gs *GameServer) onPlayerConnected(c *connection) {
	gs.Clients[c] = NewClient(c, gs)

	if gs.gameState != nil && gs.IsRunning {
		gs.SendInitialState(c)
	}
}

func (gs *GameServer) CanAcceptPlayer() bool {
	select {
	case id := <-gs.idStore.get:
		gs.idStore.free <- id
		return true
	default:
		return false
	}
	return false
}

type idStore struct {
	get  <-chan int
	free chan<- int
}

func createIdStore() *idStore {
	get := make(chan int, len(PlayerColors))
	free := make(chan int)

	res := &idStore{get: get, free: free}

	go func() {
		for i := 0; i < len(PlayerColors); i++ {
			get <- i
		}
		for {
			select {
			case id := <-free:
				get <- id
			}
		}
	}()
	return res
}

func (ids *idStore) TryGet() int {
	select {
	case id := <-ids.get:
		return id
	default:
		return -1
	}
	return -1
}
