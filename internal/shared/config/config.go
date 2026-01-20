package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port   string
	APIKey string
}

func init() {
	viper.AutomaticEnv()
}

func NewConfig() *Config {
	port := viper.GetString("PORT")
	apiKey := viper.GetString("PERSPECTIVE_API_KEY")

	if port == "" {
		log.Fatal("PORT environment variable missing")
	}
	if apiKey == "" {
		log.Fatal("PERSPECTIVE_API_KEY environment variable missing")
	}

	return &Config{
		Port:   port,
		APIKey: apiKey,
	}
}
