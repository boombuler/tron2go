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
