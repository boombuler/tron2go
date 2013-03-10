package main

import "encoding/json"
import "log"

type NewBlock struct {
	X        int
	Y        int
	PlayerId int
}

type GameStateData struct {
	Event   string
	Blocks  []NewBlock
	Players []player
}

type RoundData struct {
	Event  string
	Blocks []NewBlock
}

func (r *RoundData) Serialize() []byte {
	r.Event = "draw.blocks"

	result, err := json.Marshal(r)
	if err != nil {
		log.Println(err.Error())
	}
	return result
}

func (r *GameStateData) Serialize() []byte {
	r.Event = "draw.gamestate"

	result, err := json.Marshal(r)
	if err != nil {
		log.Println(err.Error())
	}
	return result
}
