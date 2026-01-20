package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    *AuthService
	jwtService *JWTService
}

func NewHandler(service *AuthService, jwtService *JWTService) *Handler {
	return &Handler{
		service:    service,
		jwtService: jwtService,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.service.Register(&req)
	if err != nil {
		switch err {
		case ErrEmailExists:
			c.JSON(http.StatusConflict, gin.H{
				"error": ErrEmailExists.Error(),
			})
		case ErrUsernameExists:
			c.JSON(http.StatusConflict, gin.H{
				"error": ErrUsernameExists.Error(),
			})
		default:
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "registration failed",
			})
		}
		return
	}

	token, err := h.jwtService.Generate(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.service.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	token, err := h.jwtService.Generate(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *Handler) Profile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	user, err := h.service.GetUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Middleware
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := h.jwtService.Validate(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func RegisterRoutes(r *gin.Engine, jwtSecret string) *Handler {
	repo := NewUserRepository()
	service := NewAuthService(repo)
	jwtService := NewJWTService(jwtSecret)
	handler := NewHandler(service, jwtService)

	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)
	r.GET("/profile", handler.AuthMiddleware(), handler.Profile)

	return handler
}
