package middleware

import (
	"errors"
	"time"

	"golang-clean-architecture/internal/logging"

	"github.com/gofiber/fiber/v2"
)

func RequestLogger(log *logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		started := time.Now()
		downstreamErr := c.Next()
		returnErr := downstreamErr
		if downstreamErr != nil {
			// Render the configured error response before recording the access log
			// so status and response size describe what the client received.
			returnErr = c.App().ErrorHandler(c, downstreamErr)
		}
		status := c.Response().StatusCode()
		if downstreamErr != nil {
			var fiberErr *fiber.Error
			if errors.As(downstreamErr, &fiberErr) {
				status = fiberErr.Code
			} else if status < fiber.StatusBadRequest {
				status = fiber.StatusInternalServerError
			}
		}

		fields := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latency_ms", time.Since(started).Milliseconds(),
			"response_bytes", len(c.Response().Body()),
		}
		if downstreamErr != nil {
			fields = append(fields, "error", downstreamErr)
		}
		requestLog := log.FromContext(c.UserContext())
		switch {
		case status >= fiber.StatusInternalServerError:
			requestLog.Errorw("http_request", fields...)
		case status == fiber.StatusTooManyRequests:
			requestLog.Warnw("http_request", fields...)
		default:
			requestLog.Infow("http_request", fields...)
		}
		return returnErr
	}
}
