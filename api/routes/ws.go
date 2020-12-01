package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
} // use default options

var websocketsLock sync.Mutex = sync.Mutex{}

type WebSocketConn struct {
	conn     *websocket.Conn
	connLock sync.Mutex
}

var websockets []*WebSocketConn = []*WebSocketConn{}

type websocketMessage struct {
	Type string      `json:"t"`
	Msg  interface{} `json:"m"`
}

func publishToWebsockets(i interface{}) {
	b, err := json.Marshal(i)
	if err != nil {
		log.Printf("encode: %v\n", err)
		return
	}

	for _, c := range websockets {
		if c != nil {
			c.connLock.Lock()
			err = c.conn.WriteMessage(websocket.TextMessage, b)
			c.connLock.Unlock()

			if err != nil {
				log.Printf("send: %v\n", err)
			}
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}

	websocketsLock.Lock()
	websockets = append(websockets, &WebSocketConn{conn: c, connLock: sync.Mutex{}})
	idx := len(websockets) - 1
	websocketsLock.Unlock()

	defer c.Close()
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
	}

	websockets[idx] = nil
}
