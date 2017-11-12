package layer

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/juju/errors"
	"github.com/pingcap/tidb/terror"
)

type Hub struct {
	clients   map[string]*user
	broadcast chan []byte
}

func (hub Hub) IsConnect(user string) bool {
	usuario, ok := hub.clients[user]
	if ok && len(usuario.clients) > 0 {
		return true
	}
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	maxMessageSize = 8
)

type Client struct {
	conn    *websocket.Conn
	session string
}

type user struct {
	session string
	clients map[*Client]bool
	friends []string
}

func (hub Hub) Send(user string, mensaje []byte) {
	if client, ok := hub.clients[user]; ok {
		for conection := range client.clients {
			conection.conn.WriteMessage(websocket.TextMessage, mensaje)
		}
	}
}

func (c user) Clean() {
	for key := range c.clients {
		err := key.conn.WriteMessage(websocket.PingMessage, make([]byte, 0))
		if err != nil {
			delete(c.clients, key)
		}
	}
}

func (hub *Hub) ServeWs(w http.ResponseWriter, r *http.Request, session string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		terror.Log(errors.Trace(err))
		return
	}
	client := &Client{conn: conn, session: session}
	if _, ok := hub.clients[session]; ok == false {
		hub.clients[session] = &user{
			session: session,
			clients: make(map[*Client]bool),
		}
	}
	hub.clients[session].clients[client] = true
	client.Listen()
}

func (c *Client) Listen() {
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
	log.Info("Session closed")
}
