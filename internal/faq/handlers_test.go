package faq

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mr1hm/go-chat-moderator/internal/auth"
	"github.com/mr1hm/go-chat-moderator/internal/moderation/mistralai"
)

func setupTestRouter() (*gin.Engine, *auth.JWTService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create mock services
	repo := newMockRepo()
	mistralClient := mistralai.NewClient("fake-key")
	service := NewRAGService(repo, mistralClient)
	service.LoadEmbeddings() // Mark as ready

	jwtService := auth.NewJWTService("test-secret")
	authHandler := auth.NewHandler(nil, jwtService)

	handler := NewHandler(service)

	// Public routes
	public := r.Group("")
	public.Use(handler.requireReady())
	public.POST("/ask", handler.Ask)
	public.GET("/faqs", handler.ListFAQs)

	// Protected routes
	protected := r.Group("")
	protected.Use(handler.requireReady(), authHandler.AuthMiddleware())
	protected.POST("/faqs", handler.CreateFAQ)
	protected.DELETE("/faqs/:id", handler.DeleteFAQ)

	return r, jwtService
}

func TestFAQHandlers_PublicRoutes_NoAuthRequired(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "GET /faqs is public",
			method:     "GET",
			path:       "/faqs",
			body:       nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST /ask is public",
			method:     "POST",
			path:       "/ask",
			body:       map[string]string{"question": "test question"},
			wantStatus: http.StatusInternalServerError, // Will fail because no real Mistral, but NOT 401
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			// Ensure it's not a 401 (auth error)
			if w.Code == http.StatusUnauthorized {
				t.Error("public route should not require auth")
			}
		})
	}
}

func TestFAQHandlers_ProtectedRoutes_RequireAuth(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{
			name:   "POST /faqs requires auth",
			method: "POST",
			path:   "/faqs",
			body:   map[string]string{"question": "Q", "answer": "A"},
		},
		{
			name:   "DELETE /faqs/:id requires auth",
			method: "DELETE",
			path:   "/faqs/123",
			body:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			// No auth header
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected 401 Unauthorized without token, got %d", w.Code)
			}
		})
	}
}

func TestFAQHandlers_ProtectedRoutes_WorkWithValidToken(t *testing.T) {
	router, jwtService := setupTestRouter()

	// Generate valid token
	token, err := jwtService.Generate("user-123", "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "POST /faqs works with token",
			method:     "POST",
			path:       "/faqs",
			body:       map[string]string{"question": "Q", "answer": "A"},
			wantStatus: http.StatusInternalServerError, // Fails on Mistral embed, but NOT 401
		},
		{
			name:       "DELETE /faqs/:id works with token",
			method:     "DELETE",
			path:       "/faqs/123",
			body:       nil,
			wantStatus: http.StatusNotFound, // FAQ doesn't exist, but NOT 401
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			// Should NOT be 401
			if w.Code == http.StatusUnauthorized {
				t.Error("should not get 401 with valid token")
			}
		})
	}
}

func TestFAQHandlers_ProtectedRoutes_RejectInvalidToken(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest("POST", "/faqs", bytes.NewReader([]byte(`{"question":"Q","answer":"A"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid token, got %d", w.Code)
	}
}
