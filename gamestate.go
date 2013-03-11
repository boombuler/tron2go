package main

type gameState struct {
	Board     [][]*Client
	IsRunning bool
}

func NewGameState() *gameState {
	result := &gameState{IsRunning: false}

	result.Board = make([][]*Client, FIELD_WIDTH)
	for i := range result.Board {
		result.Board[i] = make([]*Client, FIELD_HEIGHT)
	}

	return result
}

type roundState struct {
	server        *GameServer
	blocks        []NewBlock
	killedPlayers []*Client
}

func NewRound(gs *GameServer) *roundState {
	return &roundState{
		server:        gs,
		blocks:        make([]NewBlock, 0),
		killedPlayers: make([]*Client, 0),
	}
}

func (s *roundState) getPlayer(x, y int) *Client {
	for _, blk := range s.blocks {
		if blk.X == x && blk.Y == y {
			for _, p := range s.server.getPlayers(false) {
				if p.Id == blk.PlayerId {
					return p
				}
			}
			break
		}
	}
	return nil
}

func (s *roundState) killPlayer(p *Client) {
	if p.Alive {
		s.killedPlayers = append(s.killedPlayers, p)
		p.Alive = false
	}
}

func (s *roundState) movePlayer(p *Client) {
	if p.kind != Player {
		return
	}

	p.AcceptInput()

	switch p.Direction {
	case NONE:
		return
	case Left:
		p.X = p.X - 1
	case Right:
		p.X = p.X + 1
	case Up:
		p.Y = p.Y - 1
	case Down:
		p.Y = p.Y + 1
	}

	if p.X < 0 || p.Y < 0 || p.X >= FIELD_WIDTH || p.Y >= FIELD_HEIGHT || s.server.Board[p.X][p.Y] != nil {
		s.killPlayer(p)
		otherPlayer := s.getPlayer(p.X, p.Y)
		if otherPlayer != nil {
			s.killPlayer(otherPlayer)
		}
	} else {
		s.server.Board[p.X][p.Y] = p
		s.blocks = append(s.blocks, *(&NewBlock{X: p.X, Y: p.Y, PlayerId: p.Id}))
	}
}

func (s *roundState) complete() []byte {
	if len(s.killedPlayers) > 0 {
		players := s.server.getPlayers(true)
		if len(players) == 1 {
			players[0].Score += WINNER_SCORE
		}
		for _, p := range players {
			p.Score += SURVIVER_SCORE
		}
	}
	return (&RoundData{Blocks: s.blocks}).Serialize()
}
