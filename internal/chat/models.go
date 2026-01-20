package chat

import "time"

type Room struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID               string    `json:"id"`
	RoomID           string    `json:"room_id"`
	UserID           string    `json:"user_id"`
	Username         string    `json:"username,omitempty"`
	Content          string    `json:"content"`
	ModerationStatus string    `json:"moderation_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateRoomRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// Websocket Message Types
type WSMessage struct {
	Type    string      `json:"type"` // message, join, leave, error
	Payload interface{} `json:"payload"`
}
