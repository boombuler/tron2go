package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	SURVIVER_SCORE          uint          = 1
	WINNER_SCORE            uint          = 4
	FIELD_WIDTH             int           = 200
	FIELD_HEIGHT            int           = 200
	AUTOSTART_TIME          time.Duration = 5 * time.Second
	SOCKET_PATH             string        = "/socket/tron"
	SPEED                   time.Duration = (time.Second / 40)
	SUDDENDEATH_INC_TIME    time.Duration = (time.Second * 2)
	SUDDENDEATH_FACTOR      float64       = 0.35
	SUDDENDEATH_MIN_PLAYERS int           = 2
	ROOM_TIMEOUT            time.Duration = 5 * time.Second
)

var PlayerColors []string = []string{"#00ff00", "#ff0000", "#8888ff", "#00fff0", "#fff000", "#f000ff", "#aaffaa", "#ffaa00", "#aa00ff", "#ffaaaa", "#faaaff", "#ffffff"}

func serveConsts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-javascript; charset=utf-8")

	colorStr := ""
	for idx, color := range PlayerColors {
		if idx != 0 {
			colorStr += ","
		}
		colorStr += "'" + color + "'"
	}

	fmt.Fprintf(w, "FIELD_WIDTH = %v; FIELD_HEIGHT = %v;"+
		"WEBSOCKET_URL = 'ws://%v"+SOCKET_PATH+"';"+
		"PLAYER_COLORS=[%v]",
		FIELD_WIDTH, FIELD_HEIGHT, r.Host, colorStr)

}
