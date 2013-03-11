package main

type gameState struct {
	Board     [][]*Player
	IsRunning bool
}

func NewGameState() *gameState {
	result := &gameState{IsRunning: false}

	result.Board = make([][]*Player, FIELD_WIDTH)
	for i := range result.Board {
		result.Board[i] = make([]*Player, FIELD_HEIGHT)
	}

	return result
}
