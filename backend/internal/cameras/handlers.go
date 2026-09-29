package cameras

import (
	"database/sql"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// StreamHooks keeps an external stream router (go2rtc) in sync per camera.
// All hooks are best-effort: CRUD always succeeds, failures are only logged
// (startup SyncFromDB reconciles any drift).
type StreamHooks struct {
	Upsert func(id, rtspURL string)
	Delete func(id string)
}

// Register mounts CRUD under /api (caller already applies JWT middleware).
func Register(g fiber.Router, s *Store, hooks StreamHooks) {
	g.Get("/cameras", func(c *fiber.Ctx) error {
		list, err := s.List(false)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		out := make([]PublicCamera, 0, len(list))
		for _, cam := range list {
			out = append(out, cam.ToPublic())
		}
		return c.JSON(out)
	})
	g.Get("/cameras/:id", func(c *fiber.Ctx) error {
		cam, _, err := s.Get(c.Params("id"))
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "camera not found"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		return c.JSON(cam.ToPublic())
	})
	g.Post("/cameras", func(c *fiber.Ctx) error {
		var in UpsertInput
		if err := c.BodyParser(&in); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}
		cam, err := s.Create(in)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		syncStream(hooks, cam)
		return c.Status(201).JSON(cam.ToPublic())
	})
	g.Put("/cameras/:id", func(c *fiber.Ctx) error {
		var in UpsertInput
		if err := c.BodyParser(&in); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}
		cam, err := s.Update(c.Params("id"), in)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		syncStream(hooks, cam)
		return c.JSON(cam.ToPublic())
	})
	g.Delete("/cameras/:id", func(c *fiber.Ctx) error {
		if err := s.Delete(c.Params("id")); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		if hooks.Delete != nil {
			id := c.Params("id")
			go func() {
				defer func() { _ = recover() }()
				hooks.Delete(id)
			}()
		}
		return c.SendStatus(204)
	})
	g.Post("/cameras/:id/test-connection", func(c *fiber.Ctx) error {
		cam, _, err := s.Get(c.Params("id"))
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "camera not found"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		rtspOK := dialOK(net.JoinHostPort(cam.Host, strconv.Itoa(cam.RTSPPort)), 4*time.Second)
		onvifOK := dialOK(net.JoinHostPort(cam.Host, strconv.Itoa(cam.OnvifPort)), 4*time.Second)
		return c.JSON(fiber.Map{
			"camera_id": cam.ID,
			"host":      cam.Host,
			"rtsp":      fiber.Map{"port": cam.RTSPPort, "reachable": rtspOK},
			"onvif":     fiber.Map{"port": cam.OnvifPort, "reachable": onvifOK},
		})
	})
}

// syncStream mirrors one camera into the stream router:
// enabled + has URL -> upsert, otherwise -> delete (e.g. camera disabled).
// Runs async so CRUD latency is unaffected; failures are logged by the hook.
func syncStream(hooks StreamHooks, cam Camera) {
	if hooks.Upsert == nil && hooks.Delete == nil {
		return
	}
	go func() {
		defer func() { _ = recover() }()
		if cam.Enabled && cam.RTSPURL != "" && hooks.Upsert != nil {
			hooks.Upsert(cam.ID, cam.RTSPURL)
			return
		}
		if hooks.Delete != nil {
			hooks.Delete(cam.ID)
		}
	}()
}

func dialOK(addr string, timeout time.Duration) bool {	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
