package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"backend/internal/auth"
	"backend/internal/cameras"
	"backend/internal/config"
	dbpkg "backend/internal/db"
	"backend/internal/go2rtc"
	"backend/internal/onvif"
	"backend/internal/ptz"
	"backend/internal/recorder"
	"backend/internal/recordings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	_ "github.com/lib/pq"
)

var (
	clients = make(map[*websocket.Conn]bool)
	mu      sync.Mutex
	db      *sql.DB

	// 🔥 anti spam memory
	lastSeen   = make(map[int64]time.Time)
	lastSeenMu sync.Mutex
)

func main() {
	cfg := config.Load()
	db = dbpkg.Open(cfg)

	camStore := cameras.NewStore(db, cfg.CredentialKey)
	camStore.SeedFromYAML(cfg.CameraSeedPath)
	recStore := recordings.NewStore(db)
	mgr := recorder.NewManager(cfg, camStore, recStore)
	recorder.StartJanitor(cfg, db)
	mgr.AutoStart()

	// keep go2rtc streams in sync with DB (best effort)
	if list, err := camStore.List(true); err == nil {
		_ = go2rtc.SyncFromDB(cfg, list)
	}

	if cfg.InternalAPIToken == "" {
		log.Println("WARN: INTERNAL_API_TOKEN empty — /detection stays open (legacy). Set it in production.")
	}

	app := fiber.New()

	// ✅ CORS FIX (legacy kept; /api/cctv proxy in Next.js avoids CORS for new APIs)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Internal-Token",
	}))

	// ===== API DETECTION (legacy, service-to-service from YOLO) =====
	app.Post("/detection", auth.InternalToken(cfg.InternalAPIToken), func(c *fiber.Ctx) error {
		var data map[string]interface{}

		if err := c.BodyParser(&data); err != nil {
			return err
		}

		// 🔥 HANDLE BATCH
		if data["event"] == "person_detect_batch" {
			if detections, ok := data["detections"].([]interface{}); ok {
				for _, d := range detections {
					saveDetectionBatch(d.(map[string]interface{}))
				}
			}
		}

		broadcast(data)
		return c.SendStatus(200)
	})

	// ===== TOTAL UNIQUE TODAY (legacy) =====
	app.Get("/count", func(c *fiber.Ctx) error {
		cameraID := c.Query("camera_id")

		query := `
		SELECT COUNT(DISTINCT person_id)
		FROM detections
		WHERE created_at >= CURRENT_DATE
		AND created_at < CURRENT_DATE + INTERVAL '1 day'
		`
		var args []interface{}

		if cameraID != "" {
			query += " AND camera_id = $1"
			args = append(args, cameraID)
		}

		var count int
		err := db.QueryRow(query, args...).Scan(&count)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{"total": count})
	})

	// ===== HEATMAP (per jam) (legacy) =====
	app.Get("/heatmap", func(c *fiber.Ctx) error {
		cameraID := c.Query("camera_id")

		query := `
			SELECT EXTRACT(HOUR FROM created_at AT TIME ZONE 'Asia/Makassar') as hour,
			COUNT(DISTINCT person_id)
			FROM detections
			WHERE created_at >= CURRENT_DATE
			AND created_at < CURRENT_DATE + INTERVAL '1 day'
		`
		var args []interface{}

		if cameraID != "" {
			query += " AND camera_id = $1"
			args = append(args, cameraID)
		}

		query += " GROUP BY hour ORDER BY hour"

		rows, err := db.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := []fiber.Map{}

		for rows.Next() {
			var hour int
			var count int

			if err := rows.Scan(&hour, &count); err != nil {
				continue
			}

			result = append(result, fiber.Map{
				"hour":  hour,
				"count": count,
			})
		}

		return c.JSON(result)
	})

	// ===== CHART PER HARI (legacy) =====
	app.Get("/daily", func(c *fiber.Ctx) error {
		cameraID := c.Query("camera_id")

		query := `
			SELECT DATE(created_at AT TIME ZONE 'Asia/Makassar') as ddate,
			COUNT(DISTINCT person_id)
			FROM detections
		`
		var args []interface{}

		if cameraID != "" {
			query += " WHERE camera_id = $1"
			args = append(args, cameraID)
		}

		query += " GROUP BY ddate ORDER BY ddate"

		rows, err := db.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := []fiber.Map{}

		for rows.Next() {
			var date string
			var count int

			if err := rows.Scan(&date, &count); err != nil {
				continue
			}

			result = append(result, fiber.Map{
				"date":  date,
				"count": count,
			})
		}

		return c.JSON(result)
	})

	// ===== NEW CCTV API (JWT protected) =====
	api := app.Group("/api", auth.Middleware(cfg.JWTSecret))
	cameras.Register(api, camStore)
	ptzCtrl := ptz.NewController(camStore, onvif.NewGoClient())
	ptz.Register(api, ptzCtrl)
	recordings.Register(api, recStore, mgr, cfg, db)

	api.Get("/cameras/:id/stream-url", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if _, _, err := camStore.Get(id); err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "camera not found"})
		}
		return c.JSON(go2rtc.StreamURLs(cfg, id))
	})
	api.Post("/go2rtc/sync", func(c *fiber.Ctx) error {
		list, err := camStore.List(true)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		_ = go2rtc.SyncFromDB(cfg, list)
		camStore.Touch("")
		return c.JSON(fiber.Map{"ok": true, "streams": len(list)})
	})
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// ===== WS (legacy) =====
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		mu.Lock()
		clients[c] = true
		mu.Unlock()

		defer func() {
			mu.Lock()
			delete(clients, c)
			mu.Unlock()
			c.Close()
		}()

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	log.Fatal(app.Listen(":3001"))
}

// 🔥 ANTI SPAM INSERT (FIXED)
func saveDetectionBatch(data map[string]interface{}) {
	personVal, ok := data["person_id"].(float64)
	if !ok {
		return
	}

	personID := int64(personVal)

	cameraID := "unknown"
	if cam, ok := data["camera_id"].(string); ok {
		cameraID = cam
	}

	now := time.Now()

	lastSeenMu.Lock()
	defer lastSeenMu.Unlock()

	if last, ok := lastSeen[personID]; ok {
		if now.Sub(last) < 10*time.Second {
			return
		}
	}

	lastSeen[personID] = now

	query := `
	INSERT INTO detections (camera_id, person_id, action, position)
	VALUES ($1, $2, $3, $4)
	`

	// Ambil action & position dari data payload AI jika ada
	var act, pos interface{}
	if val, ok := data["action"].(string); ok {
		act = val
	}
	if val, ok := data["position"].(string); ok {
		pos = val
	}

	_, err := db.Exec(query, cameraID, personID, act, pos)
	if err != nil {
		log.Println("DB INSERT ERROR:", err)
	}
}

func broadcast(data interface{}) {
	mu.Lock()
	defer mu.Unlock()

	for client := range clients {
		err := client.WriteJSON(data)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
