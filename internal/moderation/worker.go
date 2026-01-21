package moderation

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mr1hm/go-chat-moderator/internal/chat"
	"github.com/mr1hm/go-chat-moderator/internal/moderation/perspective"
	"github.com/mr1hm/go-chat-moderator/internal/shared/redis"
)

const (
	queueKey       = "moderation:pending"
	toxicThreshold = 0.70
)

type Worker struct {
	perspective *perspective.Client
	messageRepo chat.MessageRepository
	logRepo     ModerationLogRepository
	ticker      *time.Ticker
}

func NewWorker(apiKey string) *Worker {
	return &Worker{
		perspective: perspective.NewClient(apiKey),
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

	var msg chat.Message
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		log.Printf("failed to unmarshal message: %v", err)
		return
	}

	// Call Perspective API
	score, err := w.perspective.Analyze(msg.Content)
	if err != nil {
		log.Printf("perspective API error: %v", err)
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
	w.messageRepo.UpdateStatus(msg.ID, status)

	// Log moderation result
	w.logRepo.Create(&ModerationLog{
		MessageID:     msg.ID,
		ToxicityScore: score,
		IsFlagged:     isFlagged,
	})

	log.Printf("Moderated message [ %s ]: score=%.2f status=%s", msg.ID, score, status)
}
