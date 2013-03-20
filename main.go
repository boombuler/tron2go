package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

var addr = flag.String("addr", ":13337", "http service address")

var contentTypes = map[string]string{
	".css":  "text/css; charset=utf-8",
	".js":   "application/x-javascript; charset=utf-8",
	".html": "text/html; charset=utf-8",
}

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
		w.Header().Set("Content-Type", contentTypes[filepath.Ext(filename)])
		_, err = w.Write(data)
	}
	if err != nil {
		http.Error(w, "Not found", 404)
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/", serveFiles)
	http.HandleFunc("/consts.js", serveConsts)
	http.HandleFunc(SOCKET_PATH, serveSocket)
	http.Handle("/rooms/", roomserver)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
