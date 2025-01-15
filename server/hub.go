package server

import (
	// "bytes"
	// "html/template"
	"encoding/json"
	"log"
)

// /define the Message
type Message struct {
	ClientID string
	Text     string
}

type WSMessage struct {
	Text    string      `json:"text"`
	Headers interface{} `json:"HEADERS`
}

type Hub struct {
	clients map[*Client]bool
	//def msgs as a slice of msg
	messages []*Message
	//we will have 3 channels
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Now write the run part, we will also be listening
func (h *Hub) Run() {
	for {
		select {
		//we are handling the reg case
		case client := <-h.register:
			h.clients[client] = true //This is not concurrent

			log.Printf("client registered %s", client.Id)

		//Handle the unreg case
		case client := <-h.unregister:
			//first check if it exists
			if _, ok := h.clients[client]; ok {
				close(client.Send)
				delete(h.clients, client)
			}

		//Handle the broadcast channel
		case msg := <-h.broadcast:
			h.messages = append(h.messages, msg)

			//broadcast this message
			for client := range h.clients {
				select {
				//This is where they send htmx to the client
				case client.Send <- messageToBytes(msg):

				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}

		}
	}

}

func messageToBytes(msg *Message) []byte {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return []byte{}
	}
	return jsonBytes
}

// For this new fxtion we will need to return a byte slice
// func getMessageTemplate(msg *Message) []byte {
// 	//Get the template & parse it, send the var message to it,  and convert d message into a byte slice
// 	tmpl, err := template.ParseFiles("template/message.html")
// 	if err != nil {
// 		log.Fatalf("template parsings: %s", err)
// 	}

// 	var renderedMessage bytes.Buffer
// 	err = tmpl.Execute(&renderedMessage, msg)
// 	if err != nil {
// 		log.Fatalf("template parsing: %s", err)
// 	}

// 	return renderedMessage.Bytes()
// }
