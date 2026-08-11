// Package logging provides the application's structured, contextual logger.
package logging

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pecut-ai/auth-service/pkg/logctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level              string
	Format             string
	Output             string
	File               string
	RotationMaxSizeMB  int
	RotationMaxBackups int
	RotationMaxAgeDays int
	RotationCompress   bool
}

// Logger wraps zap while retaining the familiar formatted methods used by the
// template's existing application layers.
type Logger struct {
	base      *zap.Logger
	component string
}

func New(cfg Config) (*Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	var encoder zapcore.Encoder
	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	case "console":
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	default:
		return nil, fmt.Errorf("unsupported LOG_FORMAT %q", cfg.Format)
	}

	var cores []zapcore.Core
	for _, output := range splitOutputs(cfg.Output) {
		switch output {
		case "stdout":
			cores = append(cores, zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level))
		case "stderr":
			cores = append(cores, zapcore.NewCore(encoder, zapcore.Lock(os.Stderr), level))
		case "file":
			if strings.TrimSpace(cfg.File) == "" {
				return nil, errors.New("LOG_FILE is required when LOG_OUTPUT includes file")
			}
			if dir := filepath.Dir(cfg.File); dir != "." {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return nil, fmt.Errorf("create log directory: %w", err)
				}
			}
			writer := zapcore.AddSync(&lumberjack.Logger{
				Filename:   cfg.File,
				MaxSize:    cfg.RotationMaxSizeMB,
				MaxBackups: cfg.RotationMaxBackups,
				MaxAge:     cfg.RotationMaxAgeDays,
				Compress:   cfg.RotationCompress,
			})
			cores = append(cores, zapcore.NewCore(encoder, writer, level))
		default:
			return nil, fmt.Errorf("unsupported LOG_OUTPUT target %q", output)
		}
	}
	if len(cores) == 0 {
		return nil, errors.New("LOG_OUTPUT must contain at least one output target")
	}

	return &Logger{base: zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))}, nil
}

func NewNop() *Logger { return &Logger{base: zap.NewNop()} }

func FromZap(base *zap.Logger) *Logger {
	if base == nil {
		base = zap.NewNop()
	}
	return &Logger{base: base}
}

func (c Config) Validate() error {
	var errs []error
	if _, err := parseLevel(c.Level); err != nil {
		errs = append(errs, err)
	}
	switch strings.ToLower(strings.TrimSpace(c.Format)) {
	case "json", "console":
	default:
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be json or console, got %q", c.Format))
	}
	outputs := splitOutputs(c.Output)
	if len(outputs) == 0 {
		errs = append(errs, errors.New("LOG_OUTPUT must contain at least one output target"))
	}
	for _, output := range outputs {
		if output != "stdout" && output != "stderr" && output != "file" {
			errs = append(errs, fmt.Errorf("unsupported LOG_OUTPUT target %q", output))
		}
		if output == "file" && strings.TrimSpace(c.File) == "" {
			errs = append(errs, errors.New("LOG_FILE is required when LOG_OUTPUT includes file"))
		}
	}
	if c.RotationMaxSizeMB <= 0 || c.RotationMaxBackups < 0 || c.RotationMaxAgeDays < 0 {
		errs = append(errs, errors.New("log rotation limits must be non-negative and max size must be greater than zero"))
	}
	return errors.Join(errs...)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{base: l.logger().With(toFields(args)...), component: l.component}
}

func (l *Logger) Component(component string) *Logger {
	return &Logger{base: l.logger().With(zap.String("component", component)), component: component}
}

func (l *Logger) WithError(err error) *Logger {
	return l.With("error", err)
}

// FromContext adds the correlation fields shared with auth-service.
func (l *Logger) FromContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}
	fields := make([]any, 0, 16)
	appendValue := func(key, value string) {
		if value != "" {
			fields = append(fields, key, value)
		}
	}
	appendValue("request_id", logctx.RequestID(ctx))
	appendValue("ip_address", logctx.IPAddress(ctx))
	appendValue("user_agent", logctx.UserAgent(ctx))
	appendValue("user_id", logctx.UserID(ctx))
	appendValue("service_id", logctx.ServiceID(ctx))
	appendValue("estate_id", logctx.EstateID(ctx))
	appendValue("commodity_id", logctx.CommodityID(ctx))
	if l.component == "" {
		appendValue("component", logctx.Component(ctx))
	}
	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}

func (l *Logger) Sync() error      { return l.logger().Sync() }
func (l *Logger) Zap() *zap.Logger { return l.logger() }

func (l *Logger) Debug(args ...any)              { l.logger().Sugar().Debug(args...) }
func (l *Logger) Debugf(msg string, args ...any) { l.logger().Sugar().Debugf(msg, args...) }
func (l *Logger) Debugw(msg string, args ...any) { l.logger().Sugar().Debugw(msg, args...) }
func (l *Logger) Info(args ...any)               { l.logger().Sugar().Info(args...) }
func (l *Logger) Infof(msg string, args ...any)  { l.logger().Sugar().Infof(msg, args...) }
func (l *Logger) Infow(msg string, args ...any)  { l.logger().Sugar().Infow(msg, args...) }
func (l *Logger) Warn(args ...any)               { l.logger().Sugar().Warn(args...) }
func (l *Logger) Warnf(msg string, args ...any)  { l.logger().Sugar().Warnf(msg, args...) }
func (l *Logger) Warnw(msg string, args ...any)  { l.logger().Sugar().Warnw(msg, args...) }
func (l *Logger) Error(args ...any)              { l.logger().Sugar().Error(args...) }
func (l *Logger) Errorf(msg string, args ...any) { l.logger().Sugar().Errorf(msg, args...) }
func (l *Logger) Errorw(msg string, args ...any) { l.logger().Sugar().Errorw(msg, args...) }
func (l *Logger) Fatal(args ...any)              { l.logger().Sugar().Fatal(args...) }
func (l *Logger) Fatalf(msg string, args ...any) { l.logger().Sugar().Fatalf(msg, args...) }

func (l *Logger) logger() *zap.Logger {
	if l == nil || l.base == nil {
		return zap.NewNop()
	}
	return l.base
}

func parseLevel(value string) (zapcore.Level, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(value)))); err != nil {
		return level, fmt.Errorf("invalid LOG_LEVEL %q: %w", value, err)
	}
	return level, nil
}

func splitOutputs(value string) []string {
	parts := strings.Split(value, ",")
	outputs := make([]string, 0, len(parts))
	for _, part := range parts {
		if output := strings.ToLower(strings.TrimSpace(part)); output != "" {
			outputs = append(outputs, output)
		}
	}
	return outputs
}

func toFields(args []any) []zap.Field {
	fields := make([]zap.Field, 0, (len(args)+1)/2)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok || strings.TrimSpace(key) == "" {
			fields = append(fields, zap.Any("invalid_log_key", args[i]))
			continue
		}
		if i+1 >= len(args) {
			fields = append(fields, zap.Any(key, nil))
			continue
		}
		if duration, ok := args[i+1].(time.Duration); ok {
			fields = append(fields, zap.Duration(key, duration))
			continue
		}
		fields = append(fields, zap.Any(key, args[i+1]))
	}
	return fields
}
