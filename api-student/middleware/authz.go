package middleware

import (
	"github.com/gofiber/fiber/v2"

	"api-student/app/model"
	"api-student/helper"
)

func GetUser(c *fiber.Ctx) (model.AuthUser, bool) {
	uid, ok1 := c.Locals(LocalsUserID).(int)
	uname, ok2 := c.Locals(LocalsUsername).(string)
	role, ok3 := c.Locals(LocalsRole).(string)

	if !ok1 || !ok2 || !ok3 {
		return model.AuthUser{}, false
	}
	return model.AuthUser{
		UserID:   uid,
		Username: uname,
		Role:     role,
	}, true
}

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := GetUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}

		if !perms.Can(user.Role, permission) {
			return helper.Fail(c, fiber.StatusForbidden,
				"role "+user.Role+" tidak memiliki hak "+permission)
		}

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		user, ok := GetUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}

		if _, granted := allowed[user.Role]; !granted {
			return helper.Fail(c, fiber.StatusForbidden,
				"role Anda tidak berhak mengakses endpoint ini")
		}

		return c.Next()
	}
}
