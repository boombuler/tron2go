package main

import (
	"bytes"
	"strconv"
)

type SuddenDeathStartData struct {
}

func (b *SuddenDeathStartData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj("")
	jw.WriteStr("Event", "draw.suddendeath")
	jw.EndObj()

	return jw.Flush()
}

type NewBlock struct {
	X        int
	Y        int
	PlayerId int
}

func (b *NewBlock) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj("")
	jw.WriteInt("X", b.X)
	jw.WriteInt("Y", b.Y)
	jw.WriteInt("PlayerId", b.PlayerId)
	jw.EndObj()

	return jw.Flush()
}

type RoundData []NewBlock

func (r *RoundData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj("")

	jw.WriteStr("Event", "draw.blocks")
	jw.StartArray("Blocks")

	for _, b := range *r {
		jw.Write("", &b)
	}

	return jw.EndArray().EndObj().Flush()
}

func (c *Client) SerializeIdentity() []byte {
	jw := new(JsonWriter)
	jw.StartObj("")
	jw.WriteStr("Event", "set.identity")
	jw.WriteInt("Id", c.Id)
	jw.WriteStr("Kind", c.kind.String())
	jw.EndObj()
	return jw.Flush()
}

func SerializeGameState(clients []Client, board [][]*Client) []byte {
	jw := new(JsonWriter).StartObj("")
	jw.WriteStr("Event", "draw.gamestate")
	jw.StartArray("Players")
	for _, p := range clients {
		jw.Write("", p)
	}
	jw.EndArray()

	buff := new(bytes.Buffer)

	for x, col := range board {
		for y, p := range col {
			if x != 0 || y != 0 {
				buff.WriteByte(',')
			}
			if p != nil {
				buff.WriteString(strconv.Itoa(p.Id))
			}
		}
	}
	jw.WriteStr("Board", string(buff.Bytes()))

	return jw.EndObj().Flush()
}

func SerializeScoreboard(clients []Client) []byte {
	jw := new(JsonWriter).StartObj("")
	jw.WriteStr("Event", "draw.scoreboard")
	jw.StartArray("Players")
	for _, p := range clients {
		jw.Write("", p)
	}
	jw.EndArray()

	return jw.EndObj().Flush()
}

type RoomData struct {
	Id         int
	MaxPlayers int
	Players    []Client
}

func (r RoomData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj("")
	jw.WriteInt("Id", r.Id)
	jw.WriteInt("MaxPlayers", r.MaxPlayers)
	jw.StartArray("Players")
	for _, p := range r.Players {
		jw.Write("", p)
	}

	return jw.EndArray().EndObj().Flush()
}

type RoomsData []RoomData

func (rooms *RoomsData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartArray("")
	for _, rd := range *rooms {
		jw.Write("", rd)
	}
	jw.EndArray()
	return jw.Flush()
}
