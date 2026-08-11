package middleware

import (
	"runtime/debug"

	"golang-clean-architecture/internal/logging"

	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
)

func Recover(log *logging.Logger) fiber.Handler {
	return fiberrecover.New(fiberrecover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, recovered any) {
			log.FromContext(c.UserContext()).Errorw(
				"http_panic_recovered",
				"panic", recovered,
				"method", c.Method(),
				"path", c.Path(),
				"stacktrace", string(debug.Stack()),
			)
		},
	})
}
