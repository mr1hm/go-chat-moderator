package moderation

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/mr1hm/go-chat-moderator/internal/chat"
	"github.com/mr1hm/go-chat-moderator/internal/moderation/mistralai"
	"github.com/mr1hm/go-chat-moderator/internal/shared/redis"
)

const (
	queueKey       = "moderation:pending"
	toxicThreshold = 0.70
	maxRetries     = 5
)

type Worker struct {
	mistralai   *mistralai.Client
	messageRepo chat.MessageRepository
	logRepo     ModerationLogRepository
	ticker      *time.Ticker
}

type QueueItem struct {
	Message    chat.Message `json:"message"`
	RetryCount int          `json:"retry_count"`
}

func NewWorker(apiKey string) *Worker {
	return &Worker{
		mistralai:   mistralai.NewClient(apiKey),
		messageRepo: chat.NewMessageRepository(),
		logRepo:     NewModerationLogRepository(),
		ticker:      time.NewTicker(time.Second),
	}
}

func (w *Worker) Run(ctx context.Context) {
	log.Println("Moderation worker started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.ticker.C:
			w.processNext(ctx)
		}
	}
}

func (w *Worker) processNext(ctx context.Context) {
	// Pop from Redis queue (block with timeout)
	result, err := redis.Client.BLPop(ctx, time.Second, queueKey).Result()
	if err != nil {
		return // Timeout or error, continue
	}

	var item QueueItem
	if err := json.Unmarshal([]byte(result[1]), &item); err != nil {
		log.Printf("failed to unmarshal queue item: %v", err)
		return
	}

	// Call MistralAI API
	score, err := w.mistralai.Analyze(item.Message.Content)
	if err != nil {
		// If rate-limited, re-queue and back off
		if strings.Contains(err.Error(), "429") {
			if item.RetryCount >= maxRetries {
				log.Printf("max retries exceeded for message [ %s ], marking as failed", item.Message.ID)
				w.messageRepo.UpdateStatus(item.Message.ID, "failed")
				return
			}

			item.RetryCount++
			log.Printf("rate limited, re-queueing message [ %s ] (attempt %d/%d)", item.Message.ID, item.RetryCount, maxRetries)

			b, err := json.Marshal(item)
			if err != nil {
				log.Printf("error while marshaling item for retry: %v", err)
				return
			}
			redis.Client.RPush(ctx, queueKey, b)
			time.Sleep(5 * time.Second)
			return
		}

		log.Printf("MistralAI API error: %v", err)
		return
	}

	// Determine status
	status := "approved"
	isFlagged := false
	if score >= toxicThreshold {
		status = "flagged"
		isFlagged = true
	}

	// Update message status
	w.messageRepo.UpdateStatus(item.Message.ID, status)

	b, err := json.Marshal(chat.WSMessage{
		Type: "moderation_update",
		Payload: map[string]string{
			"message_id": item.Message.ID,
			"status":     status,
		},
	})
	if err != nil {
		log.Printf("error while marshaling WSMessage: %v", err)
		return
	}

	redis.Client.Publish(ctx, "chat:"+item.Message.RoomID, b)

	// Log moderation result
	w.logRepo.Create(&ModerationLog{
		MessageID:     item.Message.ID,
		ToxicityScore: score,
		IsFlagged:     isFlagged,
	})

	log.Printf("Moderated message [ %s ]: score=%.2f status=%s", item.Message.ID, score, status)
}
