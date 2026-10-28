package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-student/app/repository"
	"api-student/app/service"
	"api-student/config"
	"api-student/database"
	"api-student/helper"
	"api-student/route"
)

func main() {
	config.LoadEnv()

	logger, err := config.NewAppLogger()
	if err != nil {
		log.Fatalf("logger: %v", err)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		log.Fatalf("mongodb: %v", err)
	}
	// Disconnect menutup seluruh koneksi di pool milik driver. Diberi batas
	// waktu sendiri, karena context milik request sudah tidak berlaku saat
	// aplikasi berhenti.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()
	mongoDB := mongoClient.Database(config.GetEnv("MONGO_DATABASE", "praktikum_backend"))

	app := fiber.New(fiber.Config{
		AppName: "Praktikum Backend Lanjut - Pertemuan 10 (api-student)",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			pesan := "terjadi kesalahan pada server"
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				pesan = e.Message
			}
			return c.Status(status).JSON(fiber.Map{
				"success": false,
				"message": pesan,
			})
		},
	})

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			log.Fatalf("gagal membuat jwt secret: %v", err)
		}
		jwtSecret = hex.EncodeToString(buf)
		log.Println("peringatan: JWT_SECRET kosong, memakai secret acak sesi ini")
	}

	accessTTL := time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute
	tokens := helper.NewTokenIssuer(jwtSecret, accessTTL)

	maxFails := config.GetEnvInt("LOGIN_MAX_FAILS", 6)
	lockout := time.Duration(config.GetEnvInt("LOGIN_LOCKOUT_MINUTES", 15)) * time.Minute
	limiter := service.NewLoginRateLimiter(maxFails, lockout)
	limiter.Start(5 * time.Minute)
	defer limiter.Stop()

	roleRepo := repository.NewRoleRepository(pool)
	ctxLoad, cancelLoad := context.WithTimeout(context.Background(), 5*time.Second)
	rawPerms, err := roleRepo.LoadPermissions(ctxLoad)
	cancelLoad()
	if err != nil {
		log.Fatalf("role permission: %v", err)
	}
	perms := helper.NewPermissionSet(rawPerms)

	userRepo := repository.NewUserRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	authSvc := service.NewAuthService(userRepo, refreshRepo, tokens, limiter)
	userSvc := service.NewUserService(userRepo, perms)

	// Index MongoDB dibuat sekali saat aplikasi menyala. Membuat index yang
	// sudah ada tidak melakukan apa pun, sehingga aman dipanggil berulang kali.
	portfolioRepo := repository.NewPortfolioRepository(mongoDB)
	if err := portfolioRepo.EnsureIndexes(context.Background()); err != nil {
		log.Fatalf("index student_portfolios: %v", err)
	}

	reviewRepo := repository.NewReviewRepository(mongoDB)
	if err := reviewRepo.EnsureIndexes(context.Background()); err != nil {
		log.Fatalf("index course_reviews: %v", err)
	}

	route.Register(app, route.Dependencies{
		Pool:                pool,
		Mongo:               mongoClient,
		JWT:                 tokens,
		Permissions:         perms,
		UserService:         userSvc,
		AuthService:         authSvc,
		PortfolioRepository: portfolioRepo,
		ReviewRepository:    reviewRepo,
	})

	app.Use(logger.RequestLogger())

	addr := config.GetEnv("APP_ADDR", ":"+config.GetEnv("APP_PORT", "3000"))
	log.Printf("server mendengarkan %s", addr)
	log.Fatal(app.Listen(addr))
}
