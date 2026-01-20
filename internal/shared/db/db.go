package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Register sqlite driver
)

var DB *sql.DB

func Init(dbPath string) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create database connection: %v", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	DB.Exec("PRAGMA journal_mode=WAL;")
	DB.Exec("PRAGMA foreign_keys=ON;")

	log.Printf("Connected to SQLite: %s", dbPath)
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
