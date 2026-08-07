package main

import (
        "database/sql"
        "log"
        "os"
        "sync"
        "time"

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
        initDB()

        app := fiber.New()

        // ✅ CORS FIX
        app.Use(cors.New(cors.Config{
                AllowOrigins: "*",
                AllowHeaders: "Origin, Content-Type, Accept",
        }))

        // ===== API DETECTION =====
        app.Post("/detection", func(c *fiber.Ctx) error {
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

        // ===== TOTAL UNIQUE TODAY =====
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

        // ===== HEATMAP (per jam) =====
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

        // ===== CHART PER HARI =====
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

        // ===== WS =====
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

func initDB() {
        connStr := "host=" + os.Getenv("DB_HOST") +
                " port=" + os.Getenv("DB_PORT") +
                " user=" + os.Getenv("DB_USER") +
                " password=" + os.Getenv("DB_PASS") +
                " dbname=" + os.Getenv("DB_NAME") +
                " sslmode=disable"

        var err error
        db, err = sql.Open("postgres", connStr)
        if err != nil {
                log.Fatal(err)
        }

        if err = db.Ping(); err != nil {
                log.Fatal("DB connection failed:", err)
        }

        log.Println("✅ DB Connected")

        createTable()
}

func createTable() {
        query := `
        CREATE TABLE IF NOT EXISTS detections (
                id SERIAL PRIMARY KEY,
                camera_id TEXT,
                person_id BIGINT,
                action TEXT,
                position TEXT,
                created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        );
        `

        _, err := db.Exec(query)
        if err != nil {
                log.Fatal("CREATE TABLE ERROR:", err)
        }
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
