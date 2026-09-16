package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-student/app/service"
	"api-student/helper"
)

type AppConfig struct {
	Fiber         *fiber.App
	DBPool        *pgxpool.Pool
	Logger        *AppLogger
	TokenIssuer   *helper.TokenIssuer
	LoginLimiter  *service.LoginRateLimiter
}
