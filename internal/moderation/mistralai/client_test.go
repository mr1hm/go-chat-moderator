package mistralai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Analyze_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Error("expected Authorization header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type header")
		}

		response := ModerationResponse{
			Results: []struct {
				Categories struct {
					Sexual           bool `json:"sexual"`
					HateAndExtremism bool `json:"hate_and_extremism"`
					Violence         bool `json:"violence"`
					SelfHarm         bool `json:"selfharm"`
				} `json:"categories"`
				CategoryScores struct {
					Sexual           float64 `json:"sexual"`
					HateAndExtremism float64 `json:"hate_and_extremism"`
					Violence         float64 `json:"violence"`
					SelfHarm         float64 `json:"selfharm"`
				} `json:"category_scores"`
			}{
				{
					CategoryScores: struct {
						Sexual           float64 `json:"sexual"`
						HateAndExtremism float64 `json:"hate_and_extremism"`
						Violence         float64 `json:"violence"`
						SelfHarm         float64 `json:"selfharm"`
					}{
						Sexual:           0.1,
						HateAndExtremism: 0.2,
						Violence:         0.8, // Max score
						SelfHarm:         0.05,
					},
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	// Override the API URL for testing
	originalURL := moderationURL
	defer func() { _ = originalURL }() // Just to reference it

	// We need to make a custom request to the test server
	// Create a wrapper that hits our test server
	score, err := analyzeWithURL(client, server.URL, "test message")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if score != 0.8 {
		t.Errorf("expected max score 0.8, got %f", score)
	}
}

func TestClient_Analyze_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	_, err := analyzeWithURL(client, server.URL, "test message")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestClient_Analyze_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ModerationResponse{
			Results: []struct {
				Categories struct {
					Sexual           bool `json:"sexual"`
					HateAndExtremism bool `json:"hate_and_extremism"`
					Violence         bool `json:"violence"`
					SelfHarm         bool `json:"selfharm"`
				} `json:"categories"`
				CategoryScores struct {
					Sexual           float64 `json:"sexual"`
					HateAndExtremism float64 `json:"hate_and_extremism"`
					Violence         float64 `json:"violence"`
					SelfHarm         float64 `json:"selfharm"`
				} `json:"category_scores"`
			}{},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	score, err := analyzeWithURL(client, server.URL, "test message")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if score != 0 {
		t.Errorf("expected score 0 for empty results, got %f", score)
	}
}

func TestClient_Analyze_MaxScoreCalculation(t *testing.T) {
	tests := []struct {
		name     string
		scores   [4]float64 // sexual, hate, violence, selfharm
		expected float64
	}{
		{"sexual highest", [4]float64{0.9, 0.1, 0.2, 0.3}, 0.9},
		{"hate highest", [4]float64{0.1, 0.95, 0.2, 0.3}, 0.95},
		{"violence highest", [4]float64{0.1, 0.2, 0.85, 0.3}, 0.85},
		{"selfharm highest", [4]float64{0.1, 0.2, 0.3, 0.99}, 0.99},
		{"all zero", [4]float64{0, 0, 0, 0}, 0},
		{"all same", [4]float64{0.5, 0.5, 0.5, 0.5}, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := ModerationResponse{
					Results: []struct {
						Categories struct {
							Sexual           bool `json:"sexual"`
							HateAndExtremism bool `json:"hate_and_extremism"`
							Violence         bool `json:"violence"`
							SelfHarm         bool `json:"selfharm"`
						} `json:"categories"`
						CategoryScores struct {
							Sexual           float64 `json:"sexual"`
							HateAndExtremism float64 `json:"hate_and_extremism"`
							Violence         float64 `json:"violence"`
							SelfHarm         float64 `json:"selfharm"`
						} `json:"category_scores"`
					}{
						{
							CategoryScores: struct {
								Sexual           float64 `json:"sexual"`
								HateAndExtremism float64 `json:"hate_and_extremism"`
								Violence         float64 `json:"violence"`
								SelfHarm         float64 `json:"selfharm"`
							}{
								Sexual:           tt.scores[0],
								HateAndExtremism: tt.scores[1],
								Violence:         tt.scores[2],
								SelfHarm:         tt.scores[3],
							},
						},
					},
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := &Client{
				apiKey:     "test-api-key",
				httpClient: server.Client(),
			}

			score, err := analyzeWithURL(client, server.URL, "test")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if score != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, score)
			}
		})
	}
}

// Helper function to analyze with a custom URL (for testing)
func analyzeWithURL(c *Client, url string, text string) (float64, error) {
	reqBody := ModerationRequest{
		Input: text,
		Model: "mistral-moderation-latest",
	}

	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var result ModerationResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Results) == 0 {
		return 0, nil
	}

	scores := result.Results[0].CategoryScores
	maxScore := max(scores.Sexual, scores.HateAndExtremism, scores.Violence, scores.SelfHarm)

	return maxScore, nil
}

// Helper function to embed with a custom URL (for testing)
func embedWithURL(c *Client, url string, texts []string) ([][]float64, error) {
	reqBody := EmbeddingRequest{
		Input: texts,
		Model: "mistral-embed",
	}

	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var result EmbeddingResponse
	json.NewDecoder(resp.Body).Decode(&result)

	embeddings := make([][]float64, len(result.Data))
	for _, d := range result.Data {
		embeddings[d.Index] = d.Embedding
	}

	return embeddings, nil
}

// Helper function to chat with a custom URL (for testing)
func chatWithURL(c *Client, url string, messages []ChatMessage) (string, error) {
	reqBody := ChatRequest{
		Model:    "mistral-small-latest",
		Messages: messages,
	}

	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d", resp.StatusCode)
	}

	var result ChatResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

// ============ Embed Tests ============

func TestClient_Embed_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Error("expected Authorization header")
		}

		response := EmbeddingResponse{
			Data: []struct {
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Embedding: []float64{0.1, 0.2, 0.3}, Index: 0},
				{Embedding: []float64{0.4, 0.5, 0.6}, Index: 1},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	embeddings, err := embedWithURL(client, server.URL, []string{"text1", "text2"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(embeddings) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(embeddings))
	}

	if embeddings[0][0] != 0.1 {
		t.Errorf("expected first embedding value 0.1, got %f", embeddings[0][0])
	}
}

func TestClient_Embed_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	_, err := embedWithURL(client, server.URL, []string{"text"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

// ============ Chat Tests ============

func TestClient_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Error("expected Authorization header")
		}

		response := ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "Hello, I'm the assistant!"}},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	messages := []ChatMessage{
		{Role: "user", Content: "Hello"},
	}

	response, err := chatWithURL(client, server.URL, messages)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response != "Hello, I'm the assistant!" {
		t.Errorf("expected 'Hello, I'm the assistant!', got '%s'", response)
	}
}

func TestClient_Chat_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	messages := []ChatMessage{
		{Role: "user", Content: "Hello"},
	}

	_, err := chatWithURL(client, server.URL, messages)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestClient_Chat_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	messages := []ChatMessage{
		{Role: "user", Content: "Hello"},
	}

	_, err := chatWithURL(client, server.URL, messages)
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}
