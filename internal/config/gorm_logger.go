package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang-clean-architecture/internal/logging"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const defaultSlowQueryThreshold = 200 * time.Millisecond

type gormLogger struct {
	log           *logging.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func NewGormLogger(log *logging.Logger) gormlogger.Interface {
	if log == nil {
		log = logging.NewNop()
	}
	return &gormLogger{log: log.Component("gorm"), level: gormlogger.Warn, slowThreshold: defaultSlowQueryThreshold}
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	next := *l
	next.level = level
	return &next
}

func (l *gormLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Info {
		l.log.FromContext(ctx).Infow("gorm_info", "detail", fmt.Sprintf(msg, args...))
	}
}

func (l *gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Warn {
		l.log.FromContext(ctx).Warnw("gorm_warning", "detail", fmt.Sprintf(msg, args...))
	}
}

func (l *gormLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Error {
		l.log.FromContext(ctx).Errorw("gorm_error", "detail", fmt.Sprintf(msg, args...))
	}
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.level >= gormlogger.Error:
		sql, rows := fc()
		l.log.FromContext(ctx).Errorw("gorm_query_error", "error", err, "elapsed_ms", elapsed.Milliseconds(), "rows", rows, "sql", sql)
	case elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		sql, rows := fc()
		l.log.FromContext(ctx).Warnw("gorm_slow_query", "elapsed_ms", elapsed.Milliseconds(), "threshold_ms", l.slowThreshold.Milliseconds(), "rows", rows, "sql", sql)
	case l.level >= gormlogger.Info:
		sql, rows := fc()
		l.log.FromContext(ctx).Infow("gorm_query", "elapsed_ms", elapsed.Milliseconds(), "rows", rows, "sql", sql)
	}
}
