package mistralai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	moderationURL = "https://api.mistral.ai/v1/moderations"
	embedURL      = "https://api.mistral.ai/v1/embeddings"
	chatURL       = "https://api.mistral.ai/v1/chat/completions"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

type ModerationRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type ModerationResponse struct {
	Results []struct {
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
	} `json:"results"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Embed(texts []string) ([][]float64, error) {
	reqBody := EmbeddingRequest{
		Input: texts,
		Model: "mistral-embed",
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, embedURL, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	embeddings := make([][]float64, len(result.Data))
	for _, d := range result.Data {
		embeddings[d.Index] = d.Embedding
	}

	return embeddings, nil
}

func (c *Client) Chat(messages []ChatMessage) (string, error) {
	reqBody := ChatRequest{
		Model:    "mistral-small-latest",
		Messages: messages,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling reqBody: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, chatURL, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("error creating new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

func (c *Client) Analyze(text string) (float64, error) {
	reqBody := ModerationRequest{
		Input: text,
		Model: "mistral-moderation-latest",
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("error while marshaling ModerationRequest: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, moderationURL, bytes.NewReader(b))
	if err != nil {
		return 0, fmt.Errorf("error while creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error while doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}

	var result ModerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("error while decoding response body: %w", err)
	}

	if len(result.Results) == 0 {
		return 0, nil
	}

	scores := result.Results[0].CategoryScores
	maxScore := max(scores.Sexual, scores.HateAndExtremism, scores.Violence, scores.SelfHarm)

	return maxScore, nil
}
