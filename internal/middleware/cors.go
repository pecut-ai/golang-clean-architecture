package middleware

import (
	"golang-clean-architecture/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS(cfg config.HTTPConfig) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type,X-Request-ID",
		ExposeHeaders:    fiber.HeaderXRequestID,
		AllowCredentials: cfg.AllowCredentials,
	})
}
