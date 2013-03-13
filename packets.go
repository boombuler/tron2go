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
	Players []Client
}

type RoundData struct {
	Event  string
	Blocks []NewBlock
}

type Identity struct {
	Event string
	Id    int
	Kind  string
}

type RoomData struct {
	Id int
	MaxPlayers int
	Players []Client
}

type RoomsData []RoomData

func (c *Client) SerializeIdentity() []byte {
	ident := new(Identity)
	ident.Event = "set.identity"
	ident.Id = c.Id
	ident.Kind = c.kind.String()

	result, err := json.Marshal(ident)
	if err != nil {
		log.Println(err.Error())
	}
	return result
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

func (rooms *RoomsData) Serialize() []byte{
	result, err := json.Marshal(rooms)
	if err != nil {
		log.Println(err.Error())
	}
	return result
}