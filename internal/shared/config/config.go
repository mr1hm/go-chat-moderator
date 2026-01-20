package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DBConfig
	RedisConfig
	ServerConfig
	JWTConfig
	PerspectiveConfig
}

// Individual service configs
type DBConfig struct {
	DBPath string
}
type RedisConfig struct {
	Addr string
}
type ServerConfig struct {
	Port string
}
type JWTConfig struct {
	Secret string
}
type PerspectiveConfig struct {
	Key string
}

func init() {
	viper.AutomaticEnv()
}

func NewConfig() *Config {
	return &Config{
		DBConfig:          LoadDBConfig(),
		RedisConfig:       LoadRedisConfig(),
		ServerConfig:      LoadServerConfig(),
		JWTConfig:         LoadJWTConfig(),
		PerspectiveConfig: LoadPerspectiveConfig(),
	}
}

func LoadDBConfig() DBConfig {
	dbPath := viper.GetString("DB_PATH")
	if dbPath == "" {
		dbPath = "data/chat.db"
	}
	return DBConfig{
		DBPath: dbPath,
	}
}
func LoadRedisConfig() RedisConfig {
	addr := viper.GetString("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return RedisConfig{
		Addr: addr,
	}
}
func LoadServerConfig() ServerConfig {
	port := viper.GetString("PORT")
	if port == "" {
		log.Fatal("PORT environment variable missing")
	}
	return ServerConfig{
		Port: port,
	}
}
func LoadJWTConfig() JWTConfig {
	secret := viper.GetString("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable missing")
	}
	return JWTConfig{
		Secret: secret,
	}
}
func LoadPerspectiveConfig() PerspectiveConfig {
	apiKey := viper.GetString("PERSPECTIVE_API_KEY")
	if apiKey == "" {
		log.Fatal("PERSPECTIVE_API_KEY environment variable missing")
	}
	return PerspectiveConfig{
		Key: apiKey,
	}
}
