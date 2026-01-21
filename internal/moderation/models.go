package moderation

import "time"

type ModerationLog struct {
	ID            string    `json:"id"`
	MessageID     string    `json:"message_id"`
	ToxicityScore float64   `json:"toxicity_score"`
	IsFlagged     bool      `json:"is_flagged"`
	ProcessedAt   time.Time `json:"processed_at"`
}
