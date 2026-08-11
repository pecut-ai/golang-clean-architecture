package config

import (
	"errors"

	"golang-clean-architecture/internal/logging"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
)

func FiberConfig(cfg *Config, log *logging.Logger) fiber.Config {
	return fiber.Config{
		AppName:                 cfg.App.Name,
		DisableStartupMessage:   cfg.App.Env == "production",
		EnableTrustedProxyCheck: len(cfg.HTTP.TrustedProxies) > 0,
		TrustedProxies:          cfg.HTTP.TrustedProxies,
		ProxyHeader:             fiber.HeaderXForwardedFor,
		ErrorHandler:            errorHandler(log),
	}
}

func errorHandler(log *logging.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		title := "Internal Server Error"
		detail := "The request could not be completed."

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			status = fiberErr.Code
			title = fiberErr.Message
			detail = fiberErr.Message
		}
		var humaErr *huma.ErrorModel
		if errors.As(err, &humaErr) {
			status = humaErr.GetStatus()
			title = humaErr.Title
			detail = humaErr.Detail
		}

		if status >= fiber.StatusInternalServerError {
			log.FromContext(c.UserContext()).Errorw("http_unhandled_error", "error", err, "status", status)
		}

		requestID := c.GetRespHeader(fiber.HeaderXRequestID)
		return c.Status(status).JSON(fiber.Map{
			"title":      title,
			"status":     status,
			"detail":     detail,
			"request_id": requestID,
		})
	}
}
