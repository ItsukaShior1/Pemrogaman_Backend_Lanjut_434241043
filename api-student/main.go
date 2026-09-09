// Package main adalah titik masuk aplikasi.
//
// main.go TIDAK berisi handler. Tugasnya hanya merakit komponen
// dalam urutan yang benar:
//   1. Muat env
//   2. Buat logger
//   3. Buat pool database
//   4. Buat instance Fiber
//   5. Daftarkan route
//   6. Jalankan server
//
// Bila ada handler di main.go, itu tanda route layer bocor ke main.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"api-student/config"
	"api-student/database"
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
		AppName: "Praktikum Backend Lanjut - Pertemuan 4 (api-student)",
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

	cfg := &config.AppConfig{
		Fiber:  app,
		DBPool: pool,
		Logger: logger,
	}

	route.Setup(app, cfg)

	port := config.GetEnv("APP_PORT", "3000")
	fmt.Println("Server berjalan di http://localhost:" + port)
	log.Fatal(app.Listen(":" + port))
}
