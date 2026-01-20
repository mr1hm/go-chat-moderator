package redis

import (
	"context"
	"log"

	goredis "github.com/redis/go-redis/v9"
)

var Client *goredis.Client

func Init(addr string) {
	Client = goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("Connected to Redis: %s", addr)
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}

	return nil
}
