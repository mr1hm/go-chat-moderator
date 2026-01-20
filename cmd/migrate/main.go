package main

import (
	"log"

	"github.com/mr1hm/go-chat-moderator/internal/shared/config"
	"github.com/mr1hm/go-chat-moderator/internal/shared/sqlite"
)

func main() {
	cfg := config.LoadDBConfig()
	sqlite.Init(cfg.DBPath)
	defer sqlite.Close()

	migrate()
	log.Println("Migration complete")
}

func migrate() {
	// Users table
	sqlite.DB.Exec(`
  		CREATE TABLE IF NOT EXISTS messages (
  			id TEXT PRIMARY KEY,
  			room_id TEXT REFERENCES rooms(id),
  			user_id TEXT REFERENCES users(id),
  			content TEXT NOT NULL,
  			moderation_status TEXT DEFAULT 'pending',
  			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)

	// Rooms table
	sqlite.DB.Exec(`
		CREATE TABLE IF NOT EXISTS rooms (
  			id TEXT PRIMARY KEY,
  			name TEXT NOT NULL,
  			created_by TEXT REFERENCES users(id),
  			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  		);
	`)

	// Messages table
	sqlite.DB.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
  			id TEXT PRIMARY KEY,
  			room_id TEXT REFERENCES rooms(id),
  			user_id TEXT REFERENCES users(id),
  			content TEXT NOT NULL,
  			moderation_status TEXT DEFAULT 'pending',
  			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  		);
	`)

	// Moderation logs table
	sqlite.DB.Exec(`
		CREATE TABLE IF NOT EXISTS moderation_logs (
  			id TEXT PRIMARY KEY,
  			message_id TEXT REFERENCES messages(id),
  			toxicity_score REAL,
  			is_flagged INTEGER DEFAULT 0,
  			processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
  		);
	`)

	log.Println("Tables created successfully")
}
