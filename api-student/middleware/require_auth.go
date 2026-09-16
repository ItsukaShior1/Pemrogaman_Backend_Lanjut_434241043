package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-student/helper"
)

const (
	LocalsUserID   = "user_id"
	LocalsUsername = "username"
	LocalsRole     = "role"
)

func RequireAuth(issuer *helper.TokenIssuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			c.Set("WWW-Authenticate", `Bearer realm="api-student"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "token wajib dikirim")
		}

		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			c.Set("WWW-Authenticate", `Bearer realm="api-student"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid")
		}

		claims, err := issuer.Parse(parts[1])
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api-student"`)
			if errors.Is(err, helper.ErrTokenExpired) {
				return helper.Fail(c, fiber.StatusUnauthorized, "token kedaluwarsa")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid")
		}

		c.Locals(LocalsUserID, claims.UserID)
		c.Locals(LocalsUsername, claims.Username)
		c.Locals(LocalsRole, claims.Role)

		return c.Next()
	}
}
