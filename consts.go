package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	FIELD_WIDTH      int    = 200
	FIELD_HEIGHT     int    = 200
	BORDER_COLOR     string = "#FF0000"
	BACKGROUND_COLOR string = "#000000"
	SOCKET_PATH      string = "/socket/tron"
	SPEED                   = (time.Second / 40)
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
		"BORDER_COLOR = '%v';"+
		"BACKGROUND_COLOR = '%v';"+
		"PLAYER_COLORS=[%v]",
		FIELD_WIDTH, FIELD_HEIGHT, r.Host, BORDER_COLOR, BACKGROUND_COLOR, colorStr)

}
