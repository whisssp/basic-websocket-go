package main

import (
	"bytes"
	"html/template"
	"log"
	"sync"
)

type Hub struct {
	sync.Mutex
	clients map[*Client]bool

	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client

	messages []*Message
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		messages:   make([]*Message, 0),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client] = true
			log.Printf("client registered: %v ", client.ID)

			for _, msg := range hub.messages {
				client.send <- getMessageTemplate(msg)
			}
		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok {
				close(client.send)
				delete(hub.clients, client)
			}
		case msg := <-hub.broadcast:
			hub.messages = append(hub.messages, msg)
			for client := range hub.clients {
				select {
				case client.send <- getMessageTemplate(msg):
				default:
					close(client.send)
					delete(hub.clients, client)
				}
			}
		}

	}
}

func getMessageTemplate(msg *Message) []byte {
	tmpl, err := template.ParseFiles("templates/message.html")
	if err != nil {
		log.Fatalf("template parsing: %s", err)
	}

	var renderedMessage bytes.Buffer
	err = tmpl.Execute(&renderedMessage, msg)
	if err != nil {
		log.Fatalf("template parsing: %s", err)
	}

	return renderedMessage.Bytes()
}
