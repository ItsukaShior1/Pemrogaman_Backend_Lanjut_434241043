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
	"api-student/config"
	"api-student/helper"
	"api-student/middleware"
)

func Setup(app *fiber.App, cfg *config.AppConfig) {
	app.Use(cfg.Logger.RequestLogger())
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! (api-student)")
	})

	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(cfg))

	students := api.Group("/students", middleware.RequireJSON())
	studentRepo := repository.NewStudentRepository(cfg.DBPool)
	studentSvc := service.NewStudentService(studentRepo)

	students.Get("/", studentSvc.List)
	students.Get("/:id", studentSvc.Get)
	students.Post("/", studentSvc.Create)
	students.Put("/:id", studentSvc.Replace)
	students.Patch("/:id", studentSvc.Patch)
	students.Delete("/:id", studentSvc.Delete)

	prestasis := api.Group("/prestasis", middleware.RequireJSON())
	prestasiRepo := repository.NewPrestasiRepository(cfg.DBPool)
	prestasiSvc := service.NewPrestasiService(prestasiRepo)

	prestasis.Get("/", prestasiSvc.List)
	prestasis.Get("/:id", prestasiSvc.Get)
	prestasis.Post("/", prestasiSvc.Create)

	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})
}

func healthCheck(cfg *config.AppConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := cfg.DBPool.Ping(ctx); err != nil {
			return helper.Fail(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi",
			)
		}
		return helper.Ok(c, "server dan database berjalan", nil)
	}
}

type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.JWTManager
	Permissions *helper.PermissionSet
	UserService *service.UserService
	AuthService *service.AuthService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// --- wajib login, hak akses diperiksa per endpoint ---
	users := api.Group("/users",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions

	// Hak dapat diputuskan tanpa melihat data -> middleware.
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

	// Hak bergantung pada kepemilikan data -> diperiksa di service.
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)
}
