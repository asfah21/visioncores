package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"backend/internal/config"
	_ "github.com/lib/pq"
)

// Open connects to postgres using Config and ensures base tables exist.
// detections table is kept exactly as legacy code expects.
func Open(cfg config.Config) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	if v := os.Getenv("DATABASE_URL"); v != "" {
		connStr = v
	}
	sqldb, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err := sqldb.Ping(); err != nil {
		log.Fatal("DB connection failed:", err)
	}
	log.Println("✅ DB Connected")
	Migrate(sqldb)
	return sqldb
}

func Migrate(sqldb *sql.DB) {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS detections (
			id SERIAL PRIMARY KEY,
			camera_id TEXT,
			person_id BIGINT,
			action TEXT,
			position TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS cameras (
			id TEXT PRIMARY KEY CHECK (id ~ '^[a-zA-Z0-9_-]+$'),
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			host TEXT NOT NULL,
			rtsp_port INT NOT NULL DEFAULT 8554,
			rtsp_url TEXT NOT NULL,
			username TEXT DEFAULT '',
			password_enc TEXT DEFAULT '',
			onvif_port INT NOT NULL DEFAULT 8000,
			enabled BOOLEAN NOT NULL DEFAULT true,
			auto_record BOOLEAN NOT NULL DEFAULT false,
			ptz_profile TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS recordings (
			id SERIAL PRIMARY KEY,
			camera_id TEXT NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
			start_time TIMESTAMPTZ NOT NULL,
			end_time TIMESTAMPTZ,
			duration_sec INT DEFAULT 0,
			file_path TEXT NOT NULL UNIQUE,
			file_size BIGINT DEFAULT 0,
			format TEXT NOT NULL DEFAULT 'mp4',
			status TEXT NOT NULL DEFAULT 'recording',
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_recordings_cam_start ON recordings(camera_id, start_time);`,
		`ALTER TABLE cameras ADD COLUMN IF NOT EXISTS ptz_profile TEXT NOT NULL DEFAULT '';`,
	}
	for _, q := range stmts {
		if _, err := sqldb.Exec(q); err != nil {
			log.Fatal("MIGRATE ERROR:", err)
		}
	}
}
