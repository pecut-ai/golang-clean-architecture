package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"golang-clean-architecture/internal/logging"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Log      logging.Config
	Docs     DocsConfig
	Auth     AuthConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

type AppConfig struct {
	Name string
	Env  string
	URL  string
}

type HTTPConfig struct {
	Port             int
	AllowOrigins     string
	AllowCredentials bool
	TrustedProxies   []string
	ShutdownTimeout  time.Duration
}

type DocsConfig struct {
	Enabled bool
}

type AuthConfig struct {
	Enabled        bool
	Target         string
	ServiceID      string
	InternalSecret string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Name         string
	SSLMode      string
	PoolIdle     int
	PoolMax      int
	PoolLifetime time.Duration
}

type KafkaConfig struct {
	Enabled          bool
	BootstrapServers []string
	GroupID          string
	AutoOffsetReset  string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigType("env")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	envFile := strings.TrimSpace(os.Getenv("ENV_FILE"))
	if envFile == "" {
		envFile = ".env"
	}
	v.SetConfigFile(envFile)
	if err := v.ReadInConfig(); err != nil && !isMissingConfigFile(err) {
		return nil, fmt.Errorf("read environment file: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Name: strings.TrimSpace(v.GetString("APP_NAME")),
			Env:  strings.ToLower(strings.TrimSpace(v.GetString("APP_ENV"))),
			URL:  strings.TrimRight(strings.TrimSpace(v.GetString("APP_URL")), "/"),
		},
		HTTP: HTTPConfig{
			Port:             v.GetInt("APP_PORT"),
			AllowOrigins:     strings.TrimSpace(v.GetString("ALLOW_ORIGINS")),
			AllowCredentials: v.GetBool("ALLOW_CREDENTIALS"),
			TrustedProxies:   splitList(v.GetString("APP_TRUSTED_PROXIES")),
			ShutdownTimeout:  v.GetDuration("APP_SHUTDOWN_TIMEOUT"),
		},
		Log: logging.Config{
			Level:              strings.TrimSpace(v.GetString("LOG_LEVEL")),
			Format:             strings.TrimSpace(v.GetString("LOG_FORMAT")),
			Output:             strings.TrimSpace(v.GetString("LOG_OUTPUT")),
			File:               strings.TrimSpace(v.GetString("LOG_FILE")),
			RotationMaxSizeMB:  v.GetInt("LOG_ROTATION_MAX_SIZE_MB"),
			RotationMaxBackups: v.GetInt("LOG_ROTATION_MAX_BACKUPS"),
			RotationMaxAgeDays: v.GetInt("LOG_ROTATION_MAX_AGE_DAYS"),
			RotationCompress:   v.GetBool("LOG_ROTATION_COMPRESS"),
		},
		Docs: DocsConfig{Enabled: v.GetBool("DOCS_ENABLED")},
		Auth: AuthConfig{
			Enabled:        v.GetBool("AUTH_ENABLED"),
			Target:         strings.TrimSpace(v.GetString("AUTH_SERVICE_GRPC_TARGET")),
			ServiceID:      strings.TrimSpace(v.GetString("SERVICE_ID")),
			InternalSecret: strings.TrimSpace(v.GetString("INTERNAL_SECRET")),
		},
		Database: DatabaseConfig{
			Host:         strings.TrimSpace(v.GetString("DB_HOST")),
			Port:         v.GetInt("DB_PORT"),
			Username:     strings.TrimSpace(v.GetString("DB_USERNAME")),
			Password:     v.GetString("DB_PASSWORD"),
			Name:         strings.TrimSpace(v.GetString("DB_NAME")),
			SSLMode:      strings.TrimSpace(v.GetString("DB_SSLMODE")),
			PoolIdle:     v.GetInt("DB_POOL_IDLE"),
			PoolMax:      v.GetInt("DB_POOL_MAX"),
			PoolLifetime: v.GetDuration("DB_POOL_LIFETIME"),
		},
		Kafka: KafkaConfig{
			Enabled:          v.GetBool("KAFKA_PRODUCER_ENABLED"),
			BootstrapServers: splitList(v.GetString("KAFKA_BOOTSTRAP_SERVERS")),
			GroupID:          strings.TrimSpace(v.GetString("KAFKA_GROUP_ID")),
			AutoOffsetReset:  strings.ToLower(strings.TrimSpace(v.GetString("KAFKA_AUTO_OFFSET_RESET"))),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate configuration: %w", err)
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	var errs []error
	required := func(key, value string) {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, fmt.Errorf("%s is required", key))
		}
	}

	required("APP_NAME", c.App.Name)
	required("APP_ENV", c.App.Env)
	switch c.App.Env {
	case "development", "test", "staging", "production":
	default:
		errs = append(errs, errors.New("APP_ENV must be development, test, staging, or production"))
	}
	required("APP_URL", c.App.URL)
	if parsed, err := url.ParseRequestURI(c.App.URL); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		errs = append(errs, errors.New("APP_URL must be an absolute URL"))
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		errs = append(errs, errors.New("APP_PORT must be between 1 and 65535"))
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("APP_SHUTDOWN_TIMEOUT must be greater than zero"))
	}
	if c.HTTP.AllowCredentials {
		for _, origin := range splitList(c.HTTP.AllowOrigins) {
			if origin == "*" {
				errs = append(errs, errors.New("ALLOW_ORIGINS cannot contain * when ALLOW_CREDENTIALS is true"))
				break
			}
		}
	}
	for _, proxy := range c.HTTP.TrustedProxies {
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				errs = append(errs, fmt.Errorf("APP_TRUSTED_PROXIES contains invalid IP or CIDR %q", proxy))
			}
		}
	}
	if c.Auth.Enabled {
		required("AUTH_SERVICE_GRPC_TARGET", c.Auth.Target)
		required("SERVICE_ID", c.Auth.ServiceID)
		required("INTERNAL_SECRET", c.Auth.InternalSecret)
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		errs = append(errs, errors.New("DB_PORT must be between 1 and 65535"))
	}
	required("DB_HOST", c.Database.Host)
	required("DB_USERNAME", c.Database.Username)
	required("DB_NAME", c.Database.Name)
	if c.Database.PoolIdle < 0 || c.Database.PoolMax < 1 || c.Database.PoolIdle > c.Database.PoolMax {
		errs = append(errs, errors.New("database pool limits must satisfy 0 <= DB_POOL_IDLE <= DB_POOL_MAX"))
	}
	if c.Database.PoolLifetime <= 0 {
		errs = append(errs, errors.New("DB_POOL_LIFETIME must be greater than zero"))
	}
	if c.Kafka.Enabled && len(c.Kafka.BootstrapServers) == 0 {
		errs = append(errs, errors.New("KAFKA_BOOTSTRAP_SERVERS is required when KAFKA_PRODUCER_ENABLED is true"))
	}
	if c.Kafka.AutoOffsetReset != "earliest" && c.Kafka.AutoOffsetReset != "latest" {
		errs = append(errs, errors.New("KAFKA_AUTO_OFFSET_RESET must be earliest or latest"))
	}
	if err := c.Log.Validate(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_SHUTDOWN_TIMEOUT", "10s")
	v.SetDefault("ALLOW_ORIGINS", "*")
	v.SetDefault("ALLOW_CREDENTIALS", false)
	v.SetDefault("APP_TRUSTED_PROXIES", "127.0.0.1,::1")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
	v.SetDefault("LOG_OUTPUT", "stdout")
	v.SetDefault("LOG_FILE", "logs/app.log")
	v.SetDefault("LOG_ROTATION_MAX_SIZE_MB", 50)
	v.SetDefault("LOG_ROTATION_MAX_BACKUPS", 5)
	v.SetDefault("LOG_ROTATION_MAX_AGE_DAYS", 7)
	v.SetDefault("LOG_ROTATION_COMPRESS", true)
	v.SetDefault("DOCS_ENABLED", false)
	v.SetDefault("AUTH_ENABLED", true)
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_POOL_IDLE", 10)
	v.SetDefault("DB_POOL_MAX", 100)
	v.SetDefault("DB_POOL_LIFETIME", "5m")
	v.SetDefault("KAFKA_PRODUCER_ENABLED", false)
	v.SetDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	v.SetDefault("KAFKA_GROUP_ID", "golang-clean-architecture")
	v.SetDefault("KAFKA_AUTO_OFFSET_RESET", "earliest")
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}

func isMissingConfigFile(err error) bool {
	var notFound viper.ConfigFileNotFoundError
	return errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) || os.IsNotExist(err)
}
