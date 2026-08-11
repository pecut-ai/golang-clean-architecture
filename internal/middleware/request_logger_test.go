package middleware

import (
	"net/http/httptest"
	"testing"

	"golang-clean-architecture/internal/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLoggerEmitsCorrelatedHTTPLog(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	log := logging.FromZap(zap.New(core))
	app := fiber.New()
	app.Use(requestid.New(requestid.Config{Header: fiber.HeaderXRequestID}))
	app.Use(RequestContext("api"))
	app.Use(RequestLogger(log))
	app.Get("/probe", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	req := httptest.NewRequest(fiber.MethodGet, "/probe", nil)
	req.Header.Set(fiber.HeaderXRequestID, "req-123")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	if resp.Header.Get(fiber.HeaderXRequestID) != "req-123" {
		t.Fatalf("response request id = %q", resp.Header.Get(fiber.HeaderXRequestID))
	}
	entries := observed.FilterMessage("http_request").All()
	if len(entries) != 1 {
		t.Fatalf("http_request entries = %d, want 1", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["request_id"] != "req-123" || fields["component"] != "api" || fields["path"] != "/probe" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestRequestLoggerRecordsRenderedErrorResponse(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	app := fiber.New()
	app.Use(requestid.New(requestid.Config{Header: fiber.HeaderXRequestID}))
	app.Use(RequestContext("api"))
	app.Use(RequestLogger(logging.FromZap(zap.New(core))))
	app.Get("/teapot", func(*fiber.Ctx) error { return fiber.ErrTeapot })

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/teapot", nil))
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	if resp.StatusCode != fiber.StatusTeapot || resp.ContentLength == 0 {
		t.Fatalf("response status=%d length=%d", resp.StatusCode, resp.ContentLength)
	}
	entries := observed.FilterMessage("http_request").All()
	if len(entries) != 1 || entries[0].ContextMap()["status"] != int64(fiber.StatusTeapot) {
		t.Fatalf("unexpected access log: %#v", entries)
	}
}
