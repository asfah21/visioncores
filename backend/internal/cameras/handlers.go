package cameras

import (
	"database/sql"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Register mounts CRUD under /api (caller already applies JWT middleware).
func Register(g fiber.Router, s *Store) {
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
		return c.JSON(cam.ToPublic())
	})
	g.Delete("/cameras/:id", func(c *fiber.Ctx) error {
		if err := s.Delete(c.Params("id")); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
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

func dialOK(addr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
