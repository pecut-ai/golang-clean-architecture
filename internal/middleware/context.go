package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pecut-ai/auth-service/pkg/logctx"
	authservice "github.com/pecut-ai/auth-service/pkg/v2/auth"
)

const requestIDLocal = "requestid"

func RequestContext(component string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := strings.TrimSpace(fmt.Sprint(c.Locals(requestIDLocal)))
		if requestID == "" || requestID == "<nil>" {
			requestID = strings.TrimSpace(c.Get(fiber.HeaderXRequestID))
		}
		c.Set(fiber.HeaderXRequestID, requestID)

		ctx := logctx.WithRequestFields(c.UserContext(), requestID, c.IP(), c.Get(fiber.HeaderUserAgent), component)
		c.SetUserContext(ctx)
		return c.Next()
	}
}

func AuthContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		if user, ok := authservice.GetAuthUser(ctx); ok && user != nil {
			ctx = logctx.WithUserID(ctx, user.UserID)
			c.SetUserContext(ctx)
		}
		return c.Next()
	}
}
