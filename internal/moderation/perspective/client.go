package perspective

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiURL = "https://commentanalyzer.googleapis.com/v1alpha1/comments:analyze"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type AnalyzeRequest struct {
	Comment             Comment         `json:"comment"`
	RequestedAttributes map[string]Attr `json:"requestedAttributes"`
}
type Comment struct {
	Text string `json:"text"`
}
type Attr struct{}

type AnalyzeResponse struct {
	AttributeScores map[string]Score `json:"attributeScores"`
}
type Score struct {
	SummaryScore SummaryScore `json:"summaryScore"`
}
type SummaryScore struct {
	Value float64 `json:"value"`
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
	req := AnalyzeRequest{
		Comment: Comment{Text: text},
		RequestedAttributes: map[string]Attr{
			"TOXICITY": {},
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("error while marshaling AnalyzeRequest: %w", err)
	}

	resp, err := c.httpClient.Post(fmt.Sprintf("%s?key=%s", apiURL, c.apiKey), "application/json", bytes.NewReader(b))
	if err != nil {
		return 0, fmt.Errorf("error while doing POST request: %w", err)
	}
	defer resp.Body.Close()

	var result AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("error while decoding response body: %w", err)
	}

	return result.AttributeScores["TOXICITY"].SummaryScore.Value, nil
}
