package main

import (
	"log"
	"time"
)

type GameServer struct {
	Clients    ClientMap
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
		Clients:    make(ClientMap),
		Broadcast:  make(chan []byte),
		idStore:    createIdStore(),
	}
	result.newGame()
	return result
}

func (gs *GameServer) newGame() {
	log.Println("NewGame")
	gs.gameState = NewGameState()
	for p := range gs.Clients.AllClients() {
		p.Reset(true)
	}
	gs.SendInitialState(nil)
}

func (gs *GameServer) gameLoop(endSignal <-chan bool) {
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
	ticker.Stop()
	if gs.SuddenDeathTime != nil {
		timeSinceSDStart := time.Now().Sub(*gs.SuddenDeathTime)
		SDRound := int(timeSinceSDStart / SUDDENDEATH_INC_TIME)
		speedFactor := 1.0 + (float64(SDRound) * SUDDENDEATH_FACTOR)
		speed := time.Duration(float64(SPEED.Nanoseconds()) / speedFactor)

		ticker = time.NewTicker(speed)
	} else {
		ticker = time.NewTicker(SPEED)
	}

	return ticker
}

func (gs *GameServer) calcRound() {
	players := gs.Clients.Players().AsSlice()

	if !gs.IsRunning {
		if len(players) < 2 {
			return
		}
		readyCnt := 0
		allReady := true

		for _, p := range players {
			p.AcceptInput()
			if p.Direction != NONE {
				readyCnt++
			} else {
				allReady = false
			}
		}

		if allReady {
			gs.IsRunning = true
		} else if readyCnt >= 2 {
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
	alivecount := gs.Clients.AliveCount()

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
	data := SerializeGameState(gs.Clients.AllClients(), gs.gameState.Board)

	if c != nil {
		c.send <- data
	} else if len(gs.Clients) > 0 {
		gs.Broadcast <- data
	}
}

func (gs *GameServer) SendScoreboard() {
	gs.Broadcast <- SerializeScoreboard(gs.Clients.AllClients())
}

func (gs *GameServer) SendNotification(msg string) {
	jw := new(JsonWriter)
	jw.StartObj("")
	jw.WriteStr("Event", "chat.message")
	jw.WriteStr("Message", msg)
	jw.EndObj()
	gs.Broadcast <- jw.Flush()
}

func (gs *GameServer) run(stop <-chan bool) {
	stopGameLoop := make(chan bool, 1)

	go gs.gameLoop(stopGameLoop)
	for {
		select {
		case c := <-gs.Register:
			go gs.onPlayerConnected(c)
		case c := <-gs.Unregister:
			if cl, ok := gs.Clients[c]; ok {
				gs.onPlayerDisconnected(cl)
			}
		case m := <-gs.Broadcast:
			for c, _ := range gs.Clients {
				c.send <- m
			}
		case <-stop:
			stopGameLoop <- true
			return
		}
	}
}

func (gs *GameServer) onPlayerDisconnected(cl *Client) {
	if cl.Id > 0 {
		gs.idStore.Free(cl.Id)
	}
	delete(gs.Clients, cl.conn)
	close(cl.conn.send)
	close(cl.conn.receive)
	go gs.SendScoreboard()
	go roomserver.RemoveRoomIfEmpty(gs)
	go gs.SendNotification(cl.Name + " disconnected")
}

func (gs *GameServer) onPlayerConnected(c *connection) {
	gs.Clients[c] = NewClient(c, gs)

	if gs.gameState != nil && gs.IsRunning {
		gs.SendInitialState(c)
	}
}
