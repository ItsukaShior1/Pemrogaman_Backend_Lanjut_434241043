package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"api-student/app/repository"
	"api-student/app/service"
	"api-student/helper"
	"api-student/middleware"
)

type Dependencies struct {
	Pool                *pgxpool.Pool
	Mongo               *mongo.Client
	JWT                 *helper.TokenIssuer
	Permissions         *helper.PermissionSet
	UserService         *service.UserService
	AuthService         *service.AuthService
	PortfolioRepository repository.PortfolioRepository
	ReviewRepository    repository.ReviewRepository
}

func Register(app *fiber.App, deps Dependencies) {
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! (api-student)")
	})

	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

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

	// Portofolio mahasiswa: identitas tetap di PostgreSQL, portofolionya
	// disimpan di MongoDB.
	portfolioSvc := service.NewPortfolioService(deps.PortfolioRepository, studentRepo, perms)

	students.Get("/:id/portfolio", portfolioSvc.Get)
	students.Put("/:id/portfolio", portfolioSvc.Replace)
	students.Post("/:id/portfolio/projects", portfolioSvc.AddProject)
	students.Delete("/:id/portfolio/projects/:projectId", portfolioSvc.DeleteProject)

	portfolios := api.Group("/portfolios", middleware.RequireAuth(deps.JWT))
	portfolios.Get("/stats/technologies", portfolioSvc.TechnologyStats)

	// Ulasan kelas: datanya di MongoDB, hak aksesnya tetap di PostgreSQL.
	courseRepo := repository.NewCourseRepository(deps.Pool)
	courseSvc := service.NewCourseService(courseRepo)
	reviewSvc := service.NewReviewService(deps.ReviewRepository, courseRepo, perms)

	courses := api.Group("/courses", middleware.RequireAuth(deps.JWT))
	courses.Get("/", courseSvc.List)
	courses.Get("/:id", courseSvc.Get)
	courses.Get("/:id/reviews", reviewSvc.ListByCourse)
	courses.Get("/:id/reviews/summary", reviewSvc.Summary)
	courses.Post("/:id/reviews", middleware.RequireJSON(), reviewSvc.Create)

	reviews := api.Group("/reviews",
		middleware.RequireJSON(),
		middleware.RequireAuth(deps.JWT))
	reviews.Get("/:id", reviewSvc.Get)
	reviews.Patch("/:id", reviewSvc.Update)
	reviews.Delete("/:id", reviewSvc.Delete)
	reviews.Post("/:id/replies", reviewSvc.Reply)
	reviews.Post("/:id/helpful", reviewSvc.MarkHelpful)

	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})
}

// healthCheck memeriksa DUA database. Aplikasi dianggap sehat hanya bila
// keduanya dapat dihubungi, dan pesan kegagalannya menyebut yang mana.
func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(
				c,
				fiber.StatusServiceUnavailable,
				"PostgreSQL tidak dapat dihubungi",
			)
		}
		if mongoClient != nil {
			if err := mongoClient.Ping(ctx, nil); err != nil {
				return helper.Fail(
					c,
					fiber.StatusServiceUnavailable,
					"MongoDB tidak dapat dihubungi",
				)
			}
		}
		return helper.Ok(c, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}
