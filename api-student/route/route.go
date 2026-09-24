package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-student/app/repository"
	"api-student/app/service"
	"api-student/helper"
	"api-student/middleware"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.TokenIssuer
	Permissions *helper.PermissionSet
	UserService *service.UserService
	AuthService *service.AuthService
}

func Register(app *fiber.App, deps Dependencies) {
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! (api-student)")
	})

	api := app.Group("/api/v1")
	api.Get("/health", healthCheckFromPool(deps.Pool))

	auth := api.Group("/auth", middleware.RequireJSON())
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	perms := deps.Permissions

	users := api.Group("/users",
		middleware.RequireJSON(),
		middleware.RequireAuth(deps.JWT))

	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		deps.UserService.List)
	users.Post("/",
		middleware.RequirePermission(perms, "user:update:any"),
		deps.UserService.Create)
	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		deps.UserService.Delete)
	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		deps.UserService.AssignRole)
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)

	studentRepo := repository.NewStudentRepository(deps.Pool)
	studentSvc := service.NewStudentService(studentRepo, perms)

	students := api.Group("/students",
		middleware.RequireJSON(),
		middleware.RequireAuth(deps.JWT))

	students.Get("/",
		middleware.RequirePermission(perms, "student:list"),
		studentSvc.List)
	students.Post("/",
		middleware.RequirePermission(perms, "student:create"),
		studentSvc.Create)
	students.Delete("/:id",
		middleware.RequirePermission(perms, "student:delete"),
		studentSvc.Delete)

	students.Get("/:id", studentSvc.Get)
	students.Put("/:id", studentSvc.Replace)
	students.Patch("/:id", studentSvc.Patch)

	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})
}

func healthCheckFromPool(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi",
			)
		}
		return helper.Ok(c, "server dan database berjalan", nil)
	}
}