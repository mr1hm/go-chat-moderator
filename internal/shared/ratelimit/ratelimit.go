package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/mr1hm/go-chat-moderator/internal/shared/redis"
)

func Allow(ctx context.Context, userID string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit: %s", userID)

	count, err := redis.Client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Set expiry on first request
	if count == 1 {
		redis.Client.Expire(ctx, key, window)
	}

	return count <= int64(limit), nil
}
