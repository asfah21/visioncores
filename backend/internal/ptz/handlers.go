package ptz

import (
	"backend/internal/onvif"

	"github.com/gofiber/fiber/v2"
)

func Register(g fiber.Router, c Controller) {
	g.Post("/cameras/:id/ptz/move", func(ctx *fiber.Ctx) error {
		var body struct {
			Pan      float64 `json:"pan"`
			Tilt     float64 `json:"tilt"`
			Zoom     float64 `json:"zoom"`
			Duration int     `json:"duration"`
		}
		if err := ctx.BodyParser(&body); err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}
		if err := c.Move(ctx.Context(), ctx.Params("id"), body.Pan, body.Tilt, body.Zoom, body.Duration); err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.JSON(fiber.Map{"ok": true})
	})
	g.Post("/cameras/:id/ptz/stop", func(ctx *fiber.Ctx) error {
		if err := c.Stop(ctx.Context(), ctx.Params("id")); err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.JSON(fiber.Map{"ok": true})
	})
	g.Get("/cameras/:id/ptz/status", func(ctx *fiber.Ctx) error {
		st, err := c.Status(ctx.Context(), ctx.Params("id"))
		if err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.JSON(st)
	})
	g.Get("/cameras/:id/info", func(ctx *fiber.Ctx) error {
		info, profs, caps, err := c.Info(ctx.Context(), ctx.Params("id"))
		if err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.JSON(fiber.Map{"device": info, "profiles": profs, "capabilities": caps})
	})
	g.Get("/cameras/:id/ptz/presets", func(ctx *fiber.Ctx) error {
		presets, err := c.ListPresets(ctx.Context(), ctx.Params("id"))
		if err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if presets == nil {
			presets = []onvif.Preset{}
		}
		return ctx.JSON(presets)
	})
	g.Post("/cameras/:id/ptz/goto", func(ctx *fiber.Ctx) error {
		var body struct {
			Preset string `json:"preset"`
		}
		if err := ctx.BodyParser(&body); err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}
		if err := c.GotoPreset(ctx.Context(), ctx.Params("id"), body.Preset); err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.JSON(fiber.Map{"ok": true})
	})
	g.Post("/cameras/:id/ptz/preset", func(ctx *fiber.Ctx) error {
		var body struct {
			Name string `json:"name"`
		}
		if err := ctx.BodyParser(&body); err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}
		token, err := c.SetPreset(ctx.Context(), ctx.Params("id"), body.Name)
		if err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.JSON(fiber.Map{"ok": true, "preset": token})
	})
	g.Post("/cameras/:id/ptz/home", func(ctx *fiber.Ctx) error {
		if err := c.GoHome(ctx.Context(), ctx.Params("id")); err != nil {
			return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.JSON(fiber.Map{"ok": true})
	})
}
