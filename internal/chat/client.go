package chat

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	Hub       *Hub
	Conn      *websocket.Conn
	Send      chan []byte
	UserID    string
	Username  string
	RoomID    string
	closeOnce sync.Once // Ensures channel closed exactly once
}

// close safely closes the Send channel exactly once
func (c *Client) close() {
	c.closeOnce.Do(func() {
		close(c.Send)
	})
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, username, roomID string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
		RoomID:   roomID,
	}
}

// ReadPump reads messages from WebSocket and broadcasts to hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		var req SendMessageRequest
		err := c.Conn.ReadJSON(&req)
		if err != nil {
			log.Printf("read error: %v", err)
			break
		}

		msg := &Message{
			RoomID:           c.RoomID,
			UserID:           c.UserID,
			Username:         c.Username,
			Content:          req.Content,
			ModerationStatus: "pending",
		}

		msg.ID = uuid.New().String()

		// Publish to Redis (broadcasts to all instances)
		c.Hub.PublishMessage(msg)

		go func(m *Message) {
			c.Hub.messageRepo.Create(m)
			c.Hub.QueueForModeration(m)
		}(msg)
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("error while writing to websocket: %v", err)
			return
		}
	}
}
