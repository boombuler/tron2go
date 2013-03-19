package main

import (
	"bytes"
	"strconv"
)

type SuddenDeathStartData struct {
}

func (b *SuddenDeathStartData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj()
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
	jw.StartObj()
	jw.WriteInt("X", b.X)
	jw.Next().WriteInt("Y", b.Y)
	jw.Next().WriteInt("PlayerId", b.PlayerId)
	jw.EndObj()

	return jw.Flush()
}

type RoundData []NewBlock

func (r *RoundData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj()

	jw.WriteStr("Event", "draw.blocks")
	jw.Next().WriteIdent("Blocks")
	jw.StartArray()

	for i, b := range *r {
		if i != 0 {
			jw.Next()
		}

		jw.Write(&b)
	}

	return jw.EndArray().EndObj().Flush()
}

func (c *Client) SerializeIdentity() []byte {
	jw := new(JsonWriter)
	jw.StartObj()
	jw.WriteStr("Event", "set.identity")
	jw.Next().WriteInt("Id", c.Id)
	jw.Next().WriteStr("Kind", c.kind.String())
	jw.EndObj()
	return jw.Flush()
}

func SerializeGameState(clients []Client, board [][]*Client) []byte {
	jw := new(JsonWriter).StartObj()
	jw.WriteStr("Event", "draw.gamestate").Next()
	jw.WriteIdent("Players").StartArray()
	for i, p := range clients {
		if i != 0 {
			jw.Next()
		}
		jw.Write(p)
	}
	jw.EndArray().Next()

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
	jw := new(JsonWriter).StartObj()
	jw.WriteStr("Event", "draw.scoreboard").Next()
	jw.WriteIdent("Players").StartArray()
	for i, p := range clients {
		if i != 0 {
			jw.Next()
		}
		jw.Write(p)
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
	jw.StartObj()
	jw.WriteInt("Id", r.Id).Next()
	jw.WriteInt("MaxPlayers", r.MaxPlayers).Next()
	jw.WriteIdent("Players").StartArray()
	for i, p := range r.Players {
		if i != 0 {
			jw.Next()
		}
		jw.Write(p)
	}

	return jw.EndArray().EndObj().Flush()
}

type RoomsData []RoomData

func (rooms *RoomsData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartArray()
	for i, rd := range *rooms {
		if i != 0 {
			jw.Next()
		}
		jw.Write(rd)
	}
	jw.EndArray()
	return jw.Flush()
}
