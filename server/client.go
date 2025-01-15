package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// client will have reference to the hub
type Client struct {
	Id   string
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte //declaration of channel var
}

const (
	writeWait = 10 * time.Second
	//send pings to peer with this period, < pongWait
	pingPeriod     = (pongWait * 9) / 10
	pongWait       = 60 * time.Second
	maxMessageSize = 512
)

// upgrader handles WS handshake process
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Handles incoming WS conn. requests
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	id := uuid.New().String()

	//create the client {client will be a ptr to the address}
	client := &Client{
		Id:   id,
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte),
	}

	//register recieves ptr to client
	//reg newly creates client w hub
	client.Hub.register <- client

	for _, msg := range client.Hub.messages {
		messageBytes := messageToBytes(msg)
		client.Send <- messageBytes
	}

	//2 Methods
	go client.readPump()
	go client.writePump()

}

// reads messgs from ws conn
func (c *Client) readPump() {
	defer func() {
		c.Conn.Close()        //close ws Conn
		c.Hub.unregister <- c //unreg client frm hub

	}()
	//set maxsize of mess client can send
	c.Conn.SetReadLimit(maxMessageSize)
	//Set Deadline 4 reading mess frm client
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	//Set hndlr 4 incoming pong mess frm client
	c.Conn.SetPongHandler(func(appData string) error {
		//Reads deadline based on d current time + pongWait
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		//indicates end of curr operation
		return nil
	})
	//strt lstng 4 mess in a loop
	for {
		_, text, err := c.Conn.ReadMessage()
		log.Printf("value %v", string(text))
		//handle errors
		if err != nil {
			//Checks if the error is an unexpected closure of the WebSocket Connection.
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %w", err)
			}
			break
		}

		//This will send mess frm Conn 2da hub

		//creates instance of WSMessage struct to store d decoded message
		msg := &WSMessage{}
		//converts text byte slice into reader object
		reader := bytes.NewReader(text)
		//creates JSON decoder for reader
		decoder := json.NewDecoder(reader)
		//Decode JSON data into msg obj of type WSMessage
		err = decoder.Decode(msg)
		if err != nil {
			log.Printf("error: %w", err)
		}

		//parse d mess and send to the hub
		//Send mess to the hub broadcast channel
		c.Hub.broadcast <- &Message{ClientID: c.Id, Text: msg.Text}
	}

}

// Client struct method dat handles messages to WS Conn
func (c *Client) writePump() {

	//triggers at regular intervals to keep WS Conn alive
	ticker := time.NewTicker(pingPeriod)
	//ensures WS Conn's closed when writePump exists
	defer func() {
		c.Conn.Close()
	}()

	for {
		//srt loop to cont. listen 4 events using select statement
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			//Write the msg from Send channel to the ws with w
			w.Write(msg)

			//Add the queued msg frm chat to the queried WS msg
			//compile all of dem into one big msg & just Send dem

			lnt := len(c.Send) //lenght of channel msg
			for i := 0; i < lnt; i++ {
				w.Write(msg)
			}

			//Ensures writer is closed after sending msg & exist fxtion if it occurs
			if err := w.Close(); err != nil {
				return
			}

			//The ticker code
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}
	}
}
