package main

type Message struct {
	ClientID string `json:"clientID"`
	Text     string `json:"text"`
}

type WSMessage struct {
	Text    string      `json:"text"`
	Headers interface{} `json:"headers"`
}
