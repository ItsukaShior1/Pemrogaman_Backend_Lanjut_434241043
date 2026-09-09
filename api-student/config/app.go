// Package config berisi konfigurasi terapan aplikasi: struct App
// yang membawa seluruh dependency siap pakai (config, logger, pool).
//
// Struct App digunakan oleh main.go sebagai "wadah" yang dilewatkan
// ke route.Setup(). Dengan cara ini main.go tidak perlu tahu detail
// inisialisasi tiap komponen.
package config

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AppConfig adalah struct global yang memuat konfigurasi dan resource
// bersama yang dipakai oleh seluruh lapisan.
type AppConfig struct {
	Fiber  *fiber.App
	DBPool *pgxpool.Pool
	Logger *AppLogger
}
