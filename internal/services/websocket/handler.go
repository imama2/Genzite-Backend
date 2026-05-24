package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/imama2/Genzite-Backend/internal/services/chatbot"
)

// upgrader holds the websocket upgrader configuration.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for now.
		// TODO: Implement a proper origin check.
		return true
	},
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, chatService *chatbot.ChatService, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := newClient(hub, conn, chatService)
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
