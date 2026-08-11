package config

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang-clean-architecture/internal/logging"

	"github.com/pecut-ai/auth-service/pkg/logctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	gormlogger "gorm.io/gorm/logger"
)

func TestGormLoggerCarriesComponentAndRequestID(t *testing.T) {
	core, observed := observer.New(zapcore.ErrorLevel)
	log := NewGormLogger(logging.FromZap(zap.New(core))).LogMode(gormlogger.Error)
	ctx := logctx.WithRequestID(context.Background(), "req-123")
	ctx = logctx.WithComponent(ctx, "api")

	log.Trace(ctx, time.Now(), func() (string, int64) { return "SELECT 1", 0 }, errors.New("query failed"))

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["component"] != "gorm" || fields["request_id"] != "req-123" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}
