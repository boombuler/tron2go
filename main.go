package main

import (
	"flag"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
)

var addr = flag.String("addr", ":13337", "http service address")
var roomcnt = flag.Int("maxrooms", 1, "maximum number of rooms")

func serveFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method nod allowed", 405)
		return
	}

	filename := r.URL.Path
	if filename == "/" {
		filename = "/index.html"
	}

	data, err := ioutil.ReadFile("./content" + filename)
	if err == nil {
		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filename)))
		_, err = w.Write(data)
	}
	if err != nil {
		http.Error(w, "Not found", 404)
	}
}

func main() {
	flag.Parse()
	roomserver.SetMaxRooms(*roomcnt)

	http.HandleFunc("/", serveFiles)
	http.HandleFunc("/consts.js", serveConsts)
	http.HandleFunc(SOCKET_PATH, serveSocket)
	http.Handle("/rooms/", roomserver)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
