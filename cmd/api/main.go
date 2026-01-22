package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mr1hm/go-chat-moderator/internal/auth"
	"github.com/mr1hm/go-chat-moderator/internal/chat"
	"github.com/mr1hm/go-chat-moderator/internal/shared/config"
	"github.com/mr1hm/go-chat-moderator/internal/shared/redis"
	"github.com/mr1hm/go-chat-moderator/internal/shared/sqlite"
)

func main() {
	// Load configs
	dbCfg := config.LoadDBConfig()
	redisCfg := config.LoadRedisConfig()
	srvCfg := config.LoadServerConfig()
	jwtCfg := config.LoadJWTConfig()

	// Init connections
	sqlite.Init(dbCfg.DBPath)
	defer sqlite.Close()

	redis.Init(redisCfg.Addr)
	defer redis.Close()

	// Setup router
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	authHandler := auth.RegisterRoutes(r, jwtCfg.Secret)
	hub := chat.NewHub()
	go hub.Run()

	chat.RegisterRoutes(r, hub, authHandler)

	log.Printf("API starting on %s", srvCfg.Port)
	r.Run(srvCfg.Port)
}
