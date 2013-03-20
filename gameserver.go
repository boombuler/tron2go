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
			ticker = gs.adjustSpeed(ticker)
		case <-endSignal:
			return
		}
	}
}

func (gs *GameServer) adjustSpeed(ticker *time.Ticker) *time.Ticker {
	if gs.SuddenDeathTime != nil {
		timeSinceSDStart := time.Now().Sub(*gs.SuddenDeathTime)
		SDRound := int(timeSinceSDStart / SUDDENDEATH_INC_TIME)
		speedFactor := 1.0 + (float64(SDRound) * SUDDENDEATH_FACTOR)
		speed := time.Duration(float64(SPEED.Nanoseconds()) / speedFactor)

		ticker.Stop()
		ticker = time.NewTicker(speed)
	} else {
		ticker = time.NewTicker(SPEED)
	}

	return ticker
}

func (gs *GameServer) calcRound() {
	players := gs.getPlayers(false)

	if !gs.IsRunning {
		if len(players) < 2 {
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
		return
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
	} else if alivecount <= SUDDENDEATH_MIN_PLAYERS {
		if gs.SuddenDeathTime == nil {
			sdStart := time.Now()
			gs.SuddenDeathTime = &sdStart
			gs.Broadcast <- new(SuddenDeathStartData).ToJson()
		}
	}
}

func (gs *GameServer) SendInitialState(c *connection) {
	clients := make([]Client, 0)

	for _, p := range gs.Clients {
		if p.kind == Player {
			clients = append(clients, *p)
		}
	}

	data := SerializeGameState(clients, gs.gameState.Board)

	if c != nil {
		c.send <- data
	} else if len(gs.Clients) > 0 {
		gs.Broadcast <- data
	}
}

func (gs *GameServer) SendScoreboard() {
	clients := make([]Client, 0)

	for _, p := range gs.Clients {
		if p.kind == Player {
			clients = append(clients, *p)
		}
	}
	gs.Broadcast <- SerializeScoreboard(clients)
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

func (gs *GameServer) run() {
	go gs.gameLoop(nil)
	for {
		select {
		case c := <-gs.Register:
			go gs.onPlayerConnected(c)
		case c := <-gs.Unregister:
			if cl, ok := gs.Clients[c]; ok {
				if cl.Id > 0 {
					gs.idStore.Free(gs.Clients[c].Id)
				}
				delete(gs.Clients, c)
				close(c.send)
				close(c.receive)
				go gs.SendScoreboard()
			}
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
