package recordings

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend/internal/cameras"
	"backend/internal/config"
	"github.com/gofiber/fiber/v2"
)

// RecorderIface avoids import cycle recorder<->recordings.
type RecorderIface interface {
	Start(cameraID string) error
	Stop(cameraID string) error
	IsRunning(cameraID string) bool
}

func Register(g fiber.Router, s *Store, rec RecorderIface, cfg config.Config, db *sql.DB) {
	g.Post("/cameras/:id/recordings/start", func(c *fiber.Ctx) error {
		if err := rec.Start(c.Params("id")); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"ok": true})
	})
	g.Post("/cameras/:id/recordings/stop", func(c *fiber.Ctx) error {
		if err := rec.Stop(c.Params("id")); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"ok": true})
	})
	g.Get("/cameras/:id/recordings", func(c *fiber.Ctx) error {
		var from, to *time.Time
		if v := c.Query("date"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				from = &t
				e := t.Add(24 * time.Hour)
				to = &e
			}
		}
		if v := c.Query("start"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				from = &t
			}
		}
		if v := c.Query("end"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				to = &t
			}
		}
		lim := 200
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				lim = n
			}
		}
		list, err := s.List(c.Params("id"), from, to, lim)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		// enrich with live status + segment files for timeline UI
		segs := segmentsOnDisk(cfg, c.Params("id"), from, to)
		return c.JSON(fiber.Map{"recordings": list, "segments": segs, "recording": rec.IsRunning(c.Params("id"))})
	})
	g.Get("/cameras/:id/recordings/timeline", func(c *fiber.Ctx) error {
		date := c.Query("date", time.Now().Format("2006-01-02"))
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "bad date, use YYYY-MM-DD"})
		}
		from := t
		to := t.Add(24 * time.Hour)
		list, _ := s.List(c.Params("id"), &from, &to, 500)
		segs := segmentsOnDisk(cfg, c.Params("id"), &from, &to)
		return c.JSON(fiber.Map{"date": date, "recordings": list, "segments": segs})
	})
	g.Get("/recordings/:id", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		r, err := s.Get(id)
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		r.FilePath = ""
		return c.JSON(r)
	})
	// Stream a single MP4 segment: GET /api/recordings/:id/stream?file=<cam>/YYYY/MM/DD/HH/mm-ss.mp4
	// file is validated against base dir to block traversal. id is used when
	// available; file-only mode (id 0 from timeline picks) validates camera from path.
	g.Get("/recordings/:id/stream", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		rel := c.Query("file", "")
		if id > 0 {
			r, err := s.Get(id)
			if err != nil {
				return c.Status(404).JSON(fiber.Map{"error": "not found"})
			}
			_ = r
			if rel == "" {
				return c.Status(404).JSON(fiber.Map{"error": "no segment yet"})
			}
		} else if rel == "" {
			return c.Status(404).JSON(fiber.Map{"error": "no segment yet"})
		}
		base, _ := filepath.Abs(cfg.RecordingPath)
		target := filepath.Clean(filepath.Join(base, rel))
		reb, err := filepath.Rel(base, target)
		if err != nil || strings.HasPrefix(reb, "..") || !strings.HasSuffix(strings.ToLower(target), ".mp4") {
			return c.Status(400).JSON(fiber.Map{"error": "bad file"})
		}
		if !cameras.ValidID(strings.Split(reb, string(filepath.Separator))[0]) {
			return c.Status(400).JSON(fiber.Map{"error": "bad file"})
		}
		f, err := os.Open(target)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "file not found"})
		}
		defer f.Close()
		fi, _ := f.Stat()
		c.Set("Content-Type", "video/mp4")
		c.Set("Accept-Ranges", "bytes")
		return c.SendStream(f, int(fi.Size()))
	})
	// AI events linked to a recording window
	g.Get("/recordings/:id/events", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		r, err := s.Get(id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		end := time.Now()
		if r.EndTime != nil {
			end = *r.EndTime
		}
		rows, err := db.Query(`SELECT camera_id, person_id, action, position, created_at FROM detections
			WHERE camera_id=$1 AND created_at BETWEEN $2 AND $3 ORDER BY created_at LIMIT 200`, r.CameraID, r.StartTime, end)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		defer rows.Close()
		out := []fiber.Map{}
		for rows.Next() {
			var cam string
			var pid int64
			var act, pos sql.NullString
			var ts time.Time
			if err := rows.Scan(&cam, &pid, &act, &pos, &ts); err != nil {
				continue
			}
			out = append(out, fiber.Map{"camera_id": cam, "person_id": pid, "action": act.String, "position": pos.String, "at": ts})
		}
		return c.JSON(out)
	})
	_ = http.StatusOK
}

type Segment struct {
	File  string    `json:"file"`
	Start time.Time `json:"start"`
	Size  int64     `json:"size"`
}

func segmentsOnDisk(cfg config.Config, cameraID string, from, to *time.Time) []Segment {
	base, _ := filepath.Abs(cfg.RecordingPath)
	root := filepath.Join(base, cameraID)
	var out []Segment
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".mp4") {
			return nil
		}
		mt := info.ModTime()
		if from != nil && mt.Before(*from) {
			return nil
		}
		if to != nil && !mt.Before(*to) {
			return nil
		}
		rel, _ := filepath.Rel(base, p)
		out = append(out, Segment{File: filepath.ToSlash(rel), Start: mt, Size: info.Size()})
		if len(out) > 500 {
			return filepath.SkipDir
		}
		return nil
	})
	return out
}

func firstSegment(cfg config.Config, _ int) string {
	return ""
}
