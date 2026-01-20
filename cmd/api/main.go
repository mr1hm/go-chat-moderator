package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mr1hm/go-chat-moderator/internal/shared/config"
)

func main() {
	cfg := config.NewConfig()
	r := gin.Default()

	r.Run(cfg.Port)
}
