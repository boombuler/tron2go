package main

import "log"
import "time"

type gameserver struct {
	State      *GameState
	clients    map[*connection]*player
	Register   chan *connection
	Unregister chan *connection
	broadcast  chan []byte
	idStore    *idStore
}

var GameServer *gameserver = createGameServer()

func createGameServer() *gameserver {
	result := &gameserver{
		Register:   make(chan *connection),
		Unregister: make(chan *connection),
		clients:    make(map[*connection]*player),
		broadcast:  make(chan []byte),
		idStore:    createIdStore(),
	}
	result.newGame()
	return result
}

func (gs *gameserver) newGame() {
	gs.State = createGameState()
}

func (self *gameserver) gameLoop(endSignal chan bool) {
	ticker := time.NewTicker(SPEED)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			self.calcRound()

		case <-endSignal:
			return
		}
	}
}

func (self *gameserver) calcRound() {

	roundData := &RoundData{Blocks: make([]NewBlock, 0)}

	for _, p := range self.clients {
		if p.state.alive {
			blk := self.movePlayer(p)
			if blk != nil {
				roundData.Blocks = append(roundData.Blocks, *blk)
			}

		}
	}
	self.broadcast <- roundData.Serialize()
}

func (self *gameserver) movePlayer(p *player) *NewBlock {
	p.acceptDirection()
	switch p.state.acceptedDirection {
	case NONE:
		return nil
	case Left:
		p.state.X = p.state.X - 1
	case Right:
		p.state.X = p.state.X + 1
	case Up:
		p.state.Y = p.state.Y - 1
	case Down:
		p.state.Y = p.state.Y + 1
	}
	if p.state.X < 0 || p.state.Y < 0 || p.state.X >= FIELD_WIDTH || p.state.Y >= FIELD_HEIGHT ||
		self.State.Board[p.state.X][p.state.Y] != nil {
		p.state.alive = false
		return nil
	} else {
		self.State.Board[p.state.X][p.state.Y] = p
	}

	return &NewBlock{X: p.state.X, Y: p.state.Y, PlayerId: p.id}
}

func (self *gameserver) run() {
	go self.gameLoop(nil)
	for {
		select {
		case c := <-self.Register:
			id := <-self.idStore.get
			log.Println("New Client")
			self.clients[c] = createPlayer(c, id, self)
			self.clients[c].newPlayerState()
		case c := <-self.Unregister:
			if self.clients[c].id > 0 {
				self.idStore.free <- self.clients[c].id
			}
			delete(self.clients, c)
			close(c.send)
			close(c.receive)
		case m := <-self.broadcast:
			for c, _ := range self.clients {
				c.send <- m
			}
		}
	}
}

func (self *gameserver) acceptNewPlayer() bool {
	select {
	case id := <-self.idStore.get:
		self.idStore.free <- id
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
