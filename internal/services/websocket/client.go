package websocket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/imama2/Genzite-Backend/internal/services/chatbot"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub     *Hub
	session *chatbot.Session

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// newClient creates a new client and a corresponding chat session.
func newClient(hub *Hub, conn *websocket.Conn, chatService *chatbot.ChatService) *Client {
	// TODO: Replace placeholder userID with actual user ID from auth context.
	userID := uuid.New().String()
	sendChan := make(chan []byte, 256)

	session := chatService.HandleNewConnection(userID, sendChan)

	return &Client{
		hub:     hub,
		conn:    conn,
		send:    sendChan,
		session: session,
	}
}
