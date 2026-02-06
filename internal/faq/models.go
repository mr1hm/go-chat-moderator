package faq

import "time"

type FAQ struct {
	ID        string    `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Embedding []float64 `json:"embedding"`
	CreatedAt time.Time `json:"created_at"`
}

type AskRequest struct {
	Question string `json:"question" binding:"required,min=1,max=500"`
}

type AskResponse struct {
	Answer  string   `json:"answer"`
	Sources []string `json:"sources"`
}

type CreateFAQRequest struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
}
