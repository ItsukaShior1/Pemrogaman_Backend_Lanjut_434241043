package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

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

	app := fiber.New(fiber.Config{
		AppName: "Praktikum Backend Lanjut - Pertemuan 5 (api-student)",
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

	cfg := &config.AppConfig{
		Fiber:        app,
		DBPool:       pool,
		Logger:       logger,
		TokenIssuer:  tokens,
		LoginLimiter: limiter,
	}

	route.Setup(app, cfg)

	port := config.GetEnv("APP_PORT", "3000")
	fmt.Println("Server berjalan di http://localhost:" + port)
	log.Fatal(app.Listen(":" + port))
}
