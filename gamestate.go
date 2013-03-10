package main

type GameState struct {
	Board     [][]*player
	IsRunning bool
}

func createGameState() *GameState {
	result := &GameState{IsRunning: false}

	result.Board = make([][]*player, FIELD_WIDTH)
	for i := range result.Board {
		result.Board[i] = make([]*player, FIELD_HEIGHT)
	}

	return result
}
