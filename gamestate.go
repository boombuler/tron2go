package main

type GameState struct {
	Board [][]*player
}

func createGameState() *GameState {
	result := &GameState{}

	result.Board = make([][]*player, FIELD_WIDTH)
	for i := range result.Board {
		result.Board[i] = make([]*player, FIELD_HEIGHT)
	}

	return result
}
