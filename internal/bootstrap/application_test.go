package bootstrap

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/logging"

	"github.com/gofiber/fiber/v2"
)

func TestApplicationHealthAndAuthBoundaryWithoutExternalConnections(t *testing.T) {
	cfg := &config.Config{
		App:  config.AppConfig{Name: "test-api", Env: "test", URL: "http://localhost:3000"},
		HTTP: config.HTTPConfig{Port: 3000, AllowOrigins: "*", TrustedProxies: []string{"127.0.0.1"}, ShutdownTimeout: time.Second},
		Log: logging.Config{
			Level: "info", Format: "json", Output: "stdout", File: "test.log", RotationMaxSizeMB: 1,
		},
		Auth: config.AuthConfig{Enabled: false},
		Database: config.DatabaseConfig{
			Host: "127.0.0.1", Port: 5432, Username: "postgres", Name: "test", SSLMode: "disable", PoolIdle: 1, PoolMax: 2, PoolLifetime: time.Minute,
		},
		Kafka: config.KafkaConfig{AutoOffsetReset: "earliest"},
	}
	app, err := NewApplication(context.Background(), cfg, logging.NewNop())
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Shutdown(context.Background()) })

	healthReq := httptest.NewRequest(fiber.MethodGet, "/api/health", nil)
	healthReq.Header.Set(fiber.HeaderXRequestID, "health-123")
	healthResp, err := app.App.Test(healthReq)
	if err != nil {
		t.Fatalf("health request error = %v", err)
	}
	if healthResp.StatusCode != fiber.StatusOK || healthResp.Header.Get(fiber.HeaderXRequestID) != "health-123" {
		t.Fatalf("health response status=%d request_id=%q", healthResp.StatusCode, healthResp.Header.Get(fiber.HeaderXRequestID))
	}

	protectedResp, err := app.App.Test(httptest.NewRequest(fiber.MethodGet, "/api/me", nil))
	if err != nil {
		t.Fatalf("protected request error = %v", err)
	}
	if protectedResp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("protected status = %d, want 503", protectedResp.StatusCode)
	}
}
