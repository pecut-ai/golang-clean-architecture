package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"golang-clean-architecture/internal/logging"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenDatabase(cfg DatabaseConfig, log *logging.Logger) (*gorm.DB, error) {
	dsnURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:   cfg.Name,
	}
	query := dsnURL.Query()
	query.Set("sslmode", cfg.SSLMode)
	dsnURL.RawQuery = query.Encode()

	db, err := gorm.Open(postgres.Open(dsnURL.String()), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               NewGormLogger(log),
	})
	if err != nil {
		return nil, fmt.Errorf("open database handle: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection pool: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.PoolIdle)
	sqlDB.SetMaxOpenConns(cfg.PoolMax)
	sqlDB.SetConnMaxLifetime(cfg.PoolLifetime)
	return db, nil
}
