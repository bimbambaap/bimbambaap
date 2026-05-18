package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is niet ingesteld")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Kon niet verbinden met database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database reageert niet:", err)
	}

	DB = db
	log.Println("Database verbonden")
}

func Migrate() {
	queries := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE`,
		// Users tabel
		`CREATE TABLE IF NOT EXISTS users (
			id          SERIAL PRIMARY KEY,
			username    VARCHAR(50)  UNIQUE NOT NULL,
			email       VARCHAR(255) UNIQUE NOT NULL,
			password    VARCHAR(255) NOT NULL,
			avatar_url  TEXT DEFAULT '',
			bio         TEXT DEFAULT '',
			is_admin    BOOLEAN DEFAULT FALSE,
			created_at  TIMESTAMP DEFAULT NOW()
		)`,

		// Posts tabel
		`CREATE TABLE IF NOT EXISTS posts (
			id          SERIAL PRIMARY KEY,
			user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			caption     TEXT DEFAULT '',
			image_url   TEXT NOT NULL,
			created_at  TIMESTAMP DEFAULT NOW()
		)`,

		// Index voor snellere feed queries
		`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatal("Migratie mislukt:", err)
		}
	}

	log.Println("Database migraties klaar")
}
