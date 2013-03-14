package main

import (
	"log"
	"net/http"
)

type RoomServer struct {
	rooms []*GameServer
}

var roomserver *RoomServer = new(RoomServer)

func (rs *RoomServer) start(roomCount int) {
	if roomCount <= 0 {
		panic("There must be at least one room")
	}
	rs.rooms = make([]*GameServer, roomCount)
	for i := 0; i < roomCount; i++ {
		rs.rooms[i] = NewGameServer()
		go rs.rooms[i].run()
	}
}

func (rs *RoomServer) GetRoom(id int) *GameServer {
	if id < 0 || id >= len(rs.rooms) {
		log.Println("Invalid RoomId: ", id)
		return nil
	}
	return rs.rooms[id]
}

func (rs *RoomServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recover()
	rooms := make(RoomsData, len(rs.rooms))
	for idx, r := range rs.rooms {
		rooms[idx].Id = idx
		rooms[idx].MaxPlayers = len(PlayerColors)
		rooms[idx].Players = make([]Client, 0)

		for _, p := range r.getPlayers(false) {
			rooms[idx].Players = append(rooms[idx].Players, *p)
		}
	}
	w.Write(rooms.Serialize())
}
