package chat

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/mr1hm/go-chat-moderator/internal/shared/redis"
)

type Hub struct {
	rooms       map[string]map[*Client]bool // roomID -> clients
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *Message
	messageRepo MessageRepository
	mtx         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms:       make(map[string]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *Message),
		messageRepo: NewMessageRepository(),
	}
}

func (h *Hub) Run() {
	// Subscribe to Redis for cross-instance messaging
	go h.subscribeRedis()

	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case msg := <-h.broadcast:
			h.broadcastToRoom(msg)
		}
	}
}

func (h *Hub) addClient(client *Client) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if h.rooms[client.RoomID] == nil {
		h.rooms[client.RoomID] = make(map[*Client]bool)
	}
	h.rooms[client.RoomID][client] = true

	log.Printf("Client %s joined room %s", client.UserID, client.RoomID)
}

func (h *Hub) removeClient(client *Client) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if clients, ok := h.rooms[client.RoomID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)
			log.Printf("Client %s left room %s", client.UserID, client.RoomID)
		}
	}
}

func (h *Hub) broadcastToRoom(msg *Message) {
	h.mtx.RLock()
	clients := h.rooms[msg.RoomID]
	h.mtx.RUnlock()

	data, _ := json.Marshal(WSMessage{
		Type:    "message",
		Payload: msg,
	})

	for client := range clients {
		select {
		case client.Send <- data:
		default:
			// Client buffer full, disconnect
			h.unregister <- client
		}
	}
}

func (h *Hub) QueueForModeration(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error while marshaling message for moderation queue: %v", err)
		return
	}

	redis.Client.RPush(context.Background(), "moderation:pending", data)
}

func (h *Hub) PublishMessage(msg *Message) {
	data, _ := json.Marshal(msg)
	redis.Client.Publish(context.Background(), "chat:"+msg.RoomID, data)
}

func (h *Hub) subscribeRedis() {
	ctx := context.Background()
	pubsub := redis.Client.PSubscribe(ctx, "chat:*")
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		var message Message
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("error while unmarshaling message payload: %v", err)
			continue
		}
		h.broadcastToRoom(&message)
	}
}
