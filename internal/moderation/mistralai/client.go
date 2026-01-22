package mistralai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiURL = "https://api.mistral.ai/v1/moderations"

type Client struct {
	apiKey     string
	httpClient *http.Client
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

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
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

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(b))
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
