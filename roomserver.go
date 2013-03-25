package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RoomServer struct {
	m           sync.Mutex
	rooms       map[int]*GameServer
	stopSignals map[int]chan<- bool
	maxrooms    int
}

var roomserver *RoomServer = new(RoomServer)

func (rs *RoomServer) getFreeId() int {
	var id int = 0
	for {
		if _, ok := rs.rooms[id]; ok {
			id++
			if id >= rs.maxrooms {
				return -1
			}
		} else {
			return id
		}
	}
	return -1
}

func (rs *RoomServer) RemoveRoomIfEmpty(room *GameServer) {
	rs.m.Lock()
	defer rs.m.Unlock()

	if len(room.Clients) > 0 {
		return
	}

	for id, r := range rs.rooms {
		if r == room {
			rs.stopSignals[id] <- true
			delete(rs.rooms, id)
		}
	}
}

func (rs *RoomServer) setRoomTimeout(id int) {
	time.Sleep(ROOM_TIMEOUT)
	room, ok := rs.rooms[id]
	if ok {
		rs.RemoveRoomIfEmpty(room)
	}
}

func (rs *RoomServer) createRoom(autoclose bool) int {
	rs.m.Lock()
	defer rs.m.Unlock()

	newRoom := NewGameServer()
	var stopSignal chan bool = nil
	if autoclose {
		stopSignal = make(chan bool, 1)
	}
	go newRoom.run(stopSignal)

	id := rs.getFreeId()
	if id < 0 {
		return id
	}

	if rs.stopSignals == nil {
		rs.stopSignals = make(map[int]chan<- bool)
	}
	rs.stopSignals[id] = stopSignal
	if rs.rooms == nil {
		rs.rooms = make(map[int]*GameServer)
	}
	rs.rooms[id] = newRoom
	go rs.setRoomTimeout(id)
	return id
}

func (rs *RoomServer) SetMaxRooms(cnt int) {
	rs.maxrooms = cnt
	if cnt == 1 {
		rs.createRoom(false)
	}
}

func (rs *RoomServer) GetRoom(id int) *GameServer {
	room, ok := rs.rooms[id]
	if !ok {
		log.Println("Invalid RoomId: ", id)
		return nil
	}
	return room
}

func (rs *RoomServer) GetRoomInfo(id int) *RoomData {
	rm, ok := rs.rooms[id]
	if !ok {
		return nil
	}

	pCnt := rm.Clients.PlayerCount()
	cCnt := len(rm.Clients)
	return &RoomData{Id: id, MaxPlayers: len(PlayerColors), Players: pCnt, Spectators: (cCnt - pCnt)}
}

func (rs *RoomServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recover()
	paths := strings.Split(r.URL.Path, "/")

	if len(paths) > 0 && paths[len(paths)-1] == "new" {
		newRoom := rs.createRoom(true)
		if newRoom >= 0 {
			w.Write(rs.GetRoomInfo(newRoom).ToJson())
		} else {
			w.Write((&ClientError{Message: "Unable to create room"}).ToJson())
		}
		return
	}

	rData := &RoomsData{MaxRoomCount: rs.maxrooms}
	rData.Rooms = make([]*RoomData, 0)
	for idx, _ := range rs.rooms {
		rData.Rooms = append(rData.Rooms, rs.GetRoomInfo(idx))
	}
	w.Write(rData.ToJson())
}
