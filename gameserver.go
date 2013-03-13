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

func (gs *GameServer) getPlayers(alive bool) []*Client {
	res := make([]*Client, 0)
	for _, c := range gs.Clients {
		if c.kind == Player {
			if !alive || c.Alive {
				res = append(res, c)
			}
		}
	}
	return res
}

func (gs *GameServer) calcRound() {
	players := gs.getPlayers(false)

	if !gs.IsRunning {
		if len(players) == 0 {
			return
		}
		anyoneReady := false
		allReady := true

		for _, p := range players {
			p.AcceptInput()
			if p.Direction != NONE {
				anyoneReady = true
			} else {
				allReady = false
			}
		}

		if allReady {
			gs.IsRunning = true
		} else if anyoneReady {
			if gs.gameState.AutoStartTime == nil {
				startAt := time.Now().Add(AUTOSTART_TIME)
				gs.gameState.AutoStartTime = &startAt
			} else if time.Now().After(*(gs.gameState.AutoStartTime)) {
				log.Println("Autostart")
				for _, p := range players {
					if p.Direction == NONE {
						p.Alive = false
						p.tryElevateTo(Spectator)
					}
				}

				gs.IsRunning = true
			}
		}
		return;
	}

	roundData := NewRound(gs)

	for _, p := range players {
		if p.Alive {
			roundData.movePlayer(p)
		}
	}
	gs.Broadcast <- roundData.complete()
	gs.checkGameOver()
}

func (gs *GameServer) checkGameOver() {
	alivecount := 0
	for _, p := range gs.Clients {
		if p.kind == Player && p.Alive {
			alivecount++
		}
	}

	if alivecount < 2 {
		gs.newGame()
	}
}

func (gs *GameServer) run() {
	go gs.gameLoop(nil)
	for {
		select {
		case c := <-gs.Register:
			go gs.onPlayerConnected(c)
		case c := <-gs.Unregister:
			if gs.Clients[c].Id > 0 {
				gs.idStore.Free(gs.Clients[c].Id)
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