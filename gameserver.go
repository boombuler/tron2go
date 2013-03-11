package main

import (
    "log"
    "time"
)

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

func (self *gameserver) newGame() {
    log.Println("NewGame")
    self.State = createGameState()
    for _, p := range self.clients {
        p.newPlayerState(true)
    }
    self.sendInitialState(nil)
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

func (self *gameserver) sendInitialState(c *connection) {
    data := &GameStateData{Blocks: make([]NewBlock, 0), Players: make([]player, 0)}

    for _, p := range self.clients {
        data.Players = append(data.Players, *p)
    }

    for x, col := range self.State.Board {
        for y, p := range col {
            if p != nil {
                data.Blocks = append(data.Blocks, *&NewBlock{X: x, Y: y, PlayerId: p.Id})
            }
        }
    }
    if c != nil {
        c.send <- data.Serialize()
    } else if len(self.clients) > 0 {
        self.broadcast <- data.Serialize()
    }
}

func (self *gameserver) calcRound() {
    if !self.State.IsRunning {
        if len(self.clients) == 0 {
            return
        }
        for _, p := range self.clients {
            p.acceptDirection()
            if p.state.acceptedDirection == NONE {
                return
            }
        }
        self.State.IsRunning = true
        return
    }

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

    self.checkGameOver()
}

func (self *gameserver) checkGameOver() {
    alivecount := 0
    for _, p := range self.clients {
        if p.state.alive {
            alivecount++
        }
    }

    if alivecount < 2 {
        self.newGame()
    }
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

    return &NewBlock{X: p.state.X, Y: p.state.Y, PlayerId: p.Id}
}

func (self *gameserver) run() {
    go self.gameLoop(nil)
    for {
        select {
        case c := <-self.Register:
            go self.onPlayerConnected(c)
        case c := <-self.Unregister:
            if self.clients[c].Id > 0 {
                self.idStore.free <- self.clients[c].Id
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

func (self *gameserver) onPlayerConnected(c *connection) {
    id := <-self.idStore.get
    player := createPlayer(c, id, self)
    self.clients[c] = player
    player.newPlayerState(false)

    if self.State != nil && self.State.IsRunning {
        self.sendInitialState(c)
    } else {
        self.newGame()
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
