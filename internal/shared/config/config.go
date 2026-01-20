package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port      string
	APIKey    string
	RedisAddr string
	DBPath    string
}

func init() {
	viper.AutomaticEnv()
}

func NewConfig() *Config {
	port := viper.GetString("PORT")
	apiKey := viper.GetString("PERSPECTIVE_API_KEY")
	redisAddr := viper.GetString("REDIS_ADDR")
	dbPath := viper.GetString("DB_PATH")

	if port == "" {
		log.Fatal("PORT environment variable missing")
	}
	if apiKey == "" {
		log.Fatal("PERSPECTIVE_API_KEY environment variable missing")
	}
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR environment variable missing")
	}
	if dbPath == "" {
		log.Fatal("DB_PATH environment variable missing")
	}

	return &Config{
		Port:      port,
		APIKey:    apiKey,
		RedisAddr: redisAddr,
		DBPath:    dbPath,
	}
}
