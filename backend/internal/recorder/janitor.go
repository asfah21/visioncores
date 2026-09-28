package recorder

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"backend/internal/config"
	"database/sql"
)

// StartJanitor deletes files + rows older than retention days, once a day.
func StartJanitor(cfg config.Config, db *sql.DB) {
	if cfg.RetentionDays <= 0 {
		cfg.RetentionDays = 7
	}
	go func() {
		run(cfg, db)
		for range time.Tick(24 * time.Hour) {
			run(cfg, db)
		}
	}()
}

func run(cfg config.Config, db *sql.DB) {
	cutoff := time.Now().AddDate(0, 0, -cfg.RetentionDays)
	rows, err := db.Query(`SELECT id,file_path FROM recordings WHERE start_time < $1`, cutoff)
	if err != nil {
		return
	}
	defer rows.Close()
	type item struct {
		id   int
		path string
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.path); err == nil {
			items = append(items, it)
		}
	}
	rows.Close()
	for _, it := range items {
		// session marker lives inside day dir; delete whole day tree safely:
		// file_path = <base>/<cam>/YYYY/MM/DD/HH/session-*.txt -> remove HH dir files older than cutoff via walk below.
		dir := filepath.Dir(it.path)
		entries, _ := filepath.Glob(filepath.Join(dir, "*.mp4"))
		for _, f := range entries {
			_ = os.Remove(f)
		}
		_ = os.Remove(it.path)
		_, _ = db.Exec(`DELETE FROM recordings WHERE id=$1`, it.id)
	}
	// sweep empty dirs + stray mp4 older than cutoff under base
	base, err := filepath.Abs(cfg.RecordingPath)
	if err != nil {
		return
	}
	_ = filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Ext(p) != ".mp4" {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(p)
			log.Println("janitor removed", p)
		}
		return nil
	})
}
