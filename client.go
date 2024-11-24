package main

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type Client struct {
	ID  string
	hub *Hub

	conn *websocket.Conn
	send chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	id := uuid.New()

	client := &Client{
		ID:   id.String(),
		hub:  hub,
		conn: conn,
		send: make(chan []byte),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, text, err := c.conn.ReadMessage()
		log.Printf("value %v", string(text))
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error:%v", err)
			}
			break
		}

		msg := &WSMessage{}
		reader := bytes.NewReader(text)
		decoder := json.NewDecoder(reader)

		errD := decoder.Decode(msg)
		if errD != nil {
			log.Printf("error:: %v", errD)
		}

		c.hub.broadcast <- &Message{
			ClientID: c.ID,
			Text:     msg.Text,
		}
	}
}

func (c *Client) readPump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(msg)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(msg)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetReadDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
