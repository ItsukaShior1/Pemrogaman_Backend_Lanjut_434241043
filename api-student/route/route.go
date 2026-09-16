package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"api-student/app/repository"
	"api-student/app/service"
	"api-student/config"
	"api-student/helper"
	mw "api-student/middleware"
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

	authRepo := repository.NewUserRepository(cfg.DBPool)
	rtRepo := repository.NewRefreshTokenRepository(cfg.DBPool)
	authSvc := service.NewAuthService(authRepo, rtRepo, cfg.TokenIssuer, cfg.LoginLimiter)

	auth := api.Group("/auth")
	auth.Post("/register", authSvc.Register)
	auth.Post("/login", authSvc.Login)
	auth.Post("/refresh", authSvc.Refresh)
	auth.Post("/logout", authSvc.Logout)
	auth.Get("/me", mw.RequireAuth(cfg.TokenIssuer), authSvc.Me)

	students := api.Group("/students",
		mw.RequireJSON(),
		mw.RequireAuth(cfg.TokenIssuer),
	)
	studentRepo := repository.NewStudentRepository(cfg.DBPool)
	studentSvc := service.NewStudentService(studentRepo)

	students.Get("/", studentSvc.List)
	students.Get("/:id", studentSvc.Get)
	students.Post("/", studentSvc.Create)
	students.Put("/:id", studentSvc.Replace)
	students.Patch("/:id", studentSvc.Patch)
	students.Delete("/:id", studentSvc.Delete)

	prestasis := api.Group("/prestasis", mw.RequireJSON())
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
