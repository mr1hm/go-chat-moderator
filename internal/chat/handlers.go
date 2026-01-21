package chat

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mr1hm/go-chat-moderator/internal/auth"
)

type Handler struct {
	roomRepo    RoomRepository
	messageRepo MessageRepository
	hub         *Hub
	jwtService  *auth.JWTService
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewHandler(hub *Hub, jwtService *auth.JWTService) *Handler {
	return &Handler{
		roomRepo:    NewRoomRepository(),
		messageRepo: NewMessageRepository(),
		hub:         hub,
		jwtService:  jwtService,
	}
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")
	room := &Room{
		Name:      req.Name,
		CreatedBy: userID.(string),
	}

	if err := h.roomRepo.Create(room); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create room",
		})
		return
	}

	c.JSON(http.StatusCreated, room)
}

func (h *Handler) ListRooms(c *gin.Context) {
	rooms, err := h.roomRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list rooms",
		})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func (h *Handler) GetRoom(c *gin.Context) {
	id := c.Param("id")
	room, err := h.roomRepo.FindByID(id)
	if err != nil {
		if err == ErrRoomNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": ErrRoomNotFound.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get room",
		})
		return
	}

	c.JSON(http.StatusOK, room)
}

func (h *Handler) GetMessages(c *gin.Context) {
	roomID := c.Param("id")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	messages, err := h.messageRepo.FindByRoom(roomID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get messages for room",
		})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	roomID := c.Param("roomID")
	token := c.Query("token")

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing token",
		})
		return
	}

	claims, err := h.jwtService.Validate(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}

	// Verify room exists
	if _, err := h.roomRepo.FindByID(roomID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "room not found",
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, claims.UserID, claims.Username, roomID)
	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func RegisterRoutes(r *gin.Engine, hub *Hub, authHandler *auth.Handler) *Handler {
	handler := NewHandler(hub, authHandler.JWTService())

	rooms := r.Group("/rooms")
	rooms.Use(authHandler.AuthMiddleware())
	{
		rooms.POST("", handler.CreateRoom)
		rooms.GET("", handler.ListRooms)
		rooms.GET("/:id", handler.GetRoom)
		rooms.GET("/:id/messages", handler.GetMessages)
	}

	r.GET("/ws/:roomID", handler.HandleWebSocket)

	return handler
}
