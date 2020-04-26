package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/higordiego/pub-sub-golang/pubsub"
	uuid "github.com/satori/go.uuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func autoId() string {
	return uuid.NewV4().String()
}

var ps = &pubsub.PubSub{}

// WebSocket - handler socket
func WebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	
	client := pubsub.Client{
		ID:         autoId(),
		Connection: conn,
	}
	
	ps.AddClient(client)
	
	log.Println("New Client is connected, total: ", len(ps.Clients))

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			// this client is disconnect or error connection we do need remove subscriptions and remove client from pubsub
			ps.RemoveClient(client)
			log.Println("Total clientes and subscriptions", len(ps.Clients), len(ps.Subscriptions))
			return
		}

		ps.HandleReceiveMessage(client, messageType, p)
		 
	}
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static")
	})
	http.HandleFunc("/ws", WebSocket)

	log.Println("Server listen  http://localhost:3000")
	http.ListenAndServe(":3000", nil)
}
