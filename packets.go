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

type ClientError struct {
	Message string
}

func (ce ClientError) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj("")
	jw.WriteStr("Event", "draw.error")
	jw.WriteStr("Message", ce.Message)
	return jw.EndObj().Flush()
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
	Players    int
	Spectators int
}

func (r RoomData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj("")
	jw.WriteInt("Id", r.Id)
	jw.WriteInt("MaxPlayers", r.MaxPlayers)
	jw.WriteInt("Players", r.Players)
	jw.WriteInt("Spectators", r.Spectators)
	return jw.EndObj().Flush()
}

type RoomsData struct {
	Rooms        []*RoomData
	MaxRoomCount int
}

func (rd *RoomsData) ToJson() []byte {
	jw := new(JsonWriter)
	jw.StartObj("")
	jw.WriteInt("MaxRoomCount", rd.MaxRoomCount)
	jw.StartArray("Rooms")
	for _, room := range rd.Rooms {
		jw.Write("", room)
	}
	jw.EndArray()
	jw.EndObj()
	return jw.Flush()
}
