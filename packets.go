package main

import (
	"encoding/json"
	"log"
	"bytes"
	"strconv"
)

type NewBlock struct {
	X        int
	Y        int
	PlayerId int
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
	Id         int
	MaxPlayers int
	Players    []Client
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

func SerializeGameState(clients []Client, board [][]*Client) []byte {
	var buff bytes.Buffer

	buff.WriteString("{\"Event\":\"draw.gamestate\",\"Players\":")
	cl, _ := json.Marshal(clients)
	buff.Write([]byte(cl))
	buff.WriteString(",\"Board\":\"");

	for x, col := range board {
		for y, p := range col {
			if x != 0 || y != 0 {
				buff.WriteString(",")
			}
			if p != nil {
				buff.WriteString(strconv.Itoa(p.Id))
			}
		}
	}
	buff.WriteString("\"}")
	return buff.Bytes()
}

func (rooms *RoomsData) Serialize() []byte {
	result, err := json.Marshal(rooms)
	if err != nil {
		log.Println(err.Error())
	}
	return result
}
