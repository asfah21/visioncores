package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Claims mirrors frontend src/lib/auth.ts SignJWT payload {username, role, exp, iat}.
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Middleware verifies HS256 JWT shared with Next.js (JWT_SECRET).
// Token can come from Authorization: Bearer <jwt> (used by /api/cctv proxy)
// or from vc_session cookie (direct calls). Only /api/* uses it.
func Middleware(secret string) fiber.Handler {
	if secret == "" {
		secret = "visioncore-super-secret-key-change-in-production"
	}
	key := []byte(secret)
	return func(c *fiber.Ctx) error {
		token := ""
		if h := c.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		}
		if token == "" {
			token = c.Cookies("vc_session")
		}
		if token == "" {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		claims := &Claims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fiber.ErrUnauthorized
			}
			return key, nil
		})
		if err != nil || !parsed.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

// InternalToken guards service-to-service POST /detection from YOLO workers.
// If INTERNAL_API_TOKEN is empty, the endpoint stays open (legacy behaviour)
// but a warning is logged at startup.
func InternalToken(expected string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if expected == "" {
			return c.Next()
		}
		got := c.Get("X-Internal-Token")
		if got == "" {
			got = c.Query("token")
		}
		if got != expected {
			return c.Status(401).JSON(fiber.Map{"error": "invalid internal token"})
		}
		return c.Next()
	}
}
