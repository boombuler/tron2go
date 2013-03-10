package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":13337", "http service address")

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
		if filename == "/index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "application/x-javascript; charset=utf-8")
		}
		_, err = w.Write(data)
	}
	if err != nil {
		http.Error(w, "Not found", 404)
	}
}

func main() {
	flag.Parse()

	go GameServer.run()

	http.HandleFunc("/", serveFiles)
	http.HandleFunc("/consts.js", serveConsts)
	http.HandleFunc(SOCKET_PATH, serveSocket)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
