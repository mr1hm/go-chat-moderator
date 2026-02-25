package faq

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mr1hm/go-chat-moderator/internal/auth"
	"github.com/mr1hm/go-chat-moderator/internal/moderation/mistralai"
)

type Handler struct {
	service *RAGService
}

func NewHandler(service *RAGService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) requireReady() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.service.IsReady() {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "FAQ service initializing",
			})
			return
		}
	}
}

func (h *Handler) Ask(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.service.Ask(req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process question",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateFAQ(c *gin.Context) {
	var req CreateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	faq, err := h.service.Add(req.Question, req.Answer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create FAQ",
		})
		return
	}

	c.JSON(http.StatusCreated, faq)
}

func (h *Handler) ListFAQs(c *gin.Context) {
	faqs, err := h.service.ListFAQs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list FAQs",
		})
		return
	}

	c.JSON(http.StatusOK, faqs)
}

func (h *Handler) DeleteFAQ(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteFAQ(id); err != nil {
		if err == ErrFAQNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "FAQ not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete FAQ",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ deleted",
	})
}

func RegisterRoutes(r *gin.Engine, mistralClient *mistralai.Client, authHandler *auth.Handler) *Handler {
	repo := NewFAQRepository()
	service := NewRAGService(repo, mistralClient)

	// Load embeddings on startup
	go func() {
		if err := service.LoadEmbeddings(); err != nil {
			log.Printf("Failed to load FAQ embeddings: %v", err)
		}
		log.Println("FAQ Embeddings cache ready")
	}()

	handler := NewHandler(service)

	// Public routes (no auth required)
	public := r.Group("")
	public.Use(handler.requireReady())
	public.POST("/ask", handler.Ask)
	public.GET("/faqs", handler.ListFAQs)

	// Protected routes (auth required)
	protected := r.Group("")
	protected.Use(handler.requireReady(), authHandler.AuthMiddleware())
	protected.POST("/faqs", handler.CreateFAQ)
	protected.DELETE("/faqs/:id", handler.DeleteFAQ)

	return handler
}
