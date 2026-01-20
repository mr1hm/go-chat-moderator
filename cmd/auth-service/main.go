package main

import (
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
}
