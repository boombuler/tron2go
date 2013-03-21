package main

import (
	"bytes"
	"github.com/garyburd/go-websocket/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	// Time allowed to write a message to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next message from the client.
	readWait = 60 * time.Second

	// Send pings to client with this period. Must be less than readWait.
	pingPeriod = (readWait * 9) / 10
)

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws   *websocket.Conn
	room *GameServer

	// Buffered channel of outbound messages.
	send    chan []byte
	receive chan []byte
}

func (c *connection) cleanup() {
	c.room.Unregister <- c
	c.ws.Close()
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	defer c.cleanup()
	c.ws.SetReadDeadline(time.Now().Add(readWait))
	for {
		op, r, err := c.ws.NextReader()
		if err != nil {
			break
		}
		switch op {
		case websocket.OpPong:
			c.ws.SetReadDeadline(time.Now().Add(readWait))
		case websocket.OpText:
			message, err := ioutil.ReadAll(r)
			if err != nil {
				break
			}
			c.receive <- message
		}
	}
}

// write writes a message with the given opCode and payload.
func (c *connection) write(opCode int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(opCode, payload)
}

func writeVarInt(buff *bytes.Buffer, val uint) {
	for {
		var nextB byte = byte(val) & 0x7F
		val = val >> 7
		if val != 0 {
			nextB = nextB ^ 0x80
			buff.WriteByte(nextB)
		} else {
			buff.WriteByte(nextB)
			break
		}
	}
}

// Compress a string to a list of output symbols.
func (c *connection) compress(uncompressed []byte) []byte {
	// Build the dictionary.
	dictSize := 256
	dictionary := make(map[string]uint)
	for i := 0; i < 256; i++ {
		dictionary[string(i)] = uint(i)
	}

	w := ""
	buffer := new(bytes.Buffer)

	for _, c := range uncompressed {
		wc := w + string(c)
		if _, ok := dictionary[wc]; ok {
			w = wc
		} else {
			writeVarInt(buffer, dictionary[w])
			// Add wc to the dictionary.
			dictionary[wc] = uint(dictSize)
			dictSize++
			w = string(c)
		}
	}

	// Output the code for w.
	if w != "" {
		writeVarInt(buffer, dictionary[w])
	}
	return buffer.Bytes()
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.cleanup()
	}()

	// Buffer the sending channel otherwise other cliens could be blocked on broadcast
	buffSend := make(chan []byte)
	go BufferChannel(c.send, buffSend)

	for {
		select {
		case message, ok := <-buffSend:
			if !ok {
				c.write(websocket.OpClose, []byte{})
				return
			}
			if err := c.write(websocket.OpBinary, c.compress(message)); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.OpPing, []byte{}); err != nil {
				return
			}
		}
	}
}

// serverWs handles webocket requests from the client.
func serveSocket(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}

	ws, err := websocket.Upgrade(w, r.Header, "", 1024, 1024)

	if err != nil {
		http.Error(w, err.Error(), 400)
		log.Println(err)
		return
	}

	roomid, err := strconv.Atoi(r.URL.RawQuery)
	if err != nil {
		ws.WriteControl(websocket.OpClose, websocket.FormatCloseMessage(4000, "Not a valid room id"), time.Now().Add(writeWait))
		ws.Close()
		return
	}

	room := roomserver.GetRoom(roomid)
	if room == nil {
		ws.WriteControl(websocket.OpClose, websocket.FormatCloseMessage(4004, "Room not found"), time.Now().Add(writeWait))
		ws.Close()
		return
	}

	c := &connection{send: make(chan []byte), receive: make(chan []byte), ws: ws, room: room}
	room.Register <- c
	go c.writePump()
	c.readPump()
}
