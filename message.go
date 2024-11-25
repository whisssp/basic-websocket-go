package main

type Message struct {
	ClientID string `json:"clientID"`
	Text     string `json:"text"`
}

type WSMessage struct {
	ID      string      `json:"id"`
	Text    string      `json:"text"`
	Headers interface{} `json:"headers"`
}
