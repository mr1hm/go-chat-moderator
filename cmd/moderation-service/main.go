package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mr1hm/go-chat-moderator/internal/moderation"
	"github.com/mr1hm/go-chat-moderator/internal/shared/config"
	"github.com/mr1hm/go-chat-moderator/internal/shared/redis"
	"github.com/mr1hm/go-chat-moderator/internal/shared/sqlite"
)

func main() {
	dbCfg := config.LoadDBConfig()
	redisCfg := config.LoadRedisConfig()
	perspectiveCfg := config.LoadPerspectiveConfig()

	sqlite.Init(dbCfg.DBPath)
	defer sqlite.Close()

	redis.Init(redisCfg.Addr)
	defer redis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Moderation service shutting down...")
		cancel()
	}()

	worker := moderation.NewWorker(perspectiveCfg.Key)
	worker.Run(ctx)
}
