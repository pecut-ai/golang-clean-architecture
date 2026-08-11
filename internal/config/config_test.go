package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang-clean-architecture/internal/buildinfo"
)

func TestLoadParsesValidatedEnvironment(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("APP_PORT", "4100")
	t.Setenv("ALLOW_ORIGINS", "https://app.example.com")
	t.Setenv("DB_POOL_LIFETIME", "2m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Name != "template-api" || cfg.App.Env != "test" || cfg.App.URL != "https://api.example.com" {
		t.Fatalf("unexpected app config: %#v", cfg.App)
	}
	if cfg.HTTP.Port != 4100 || cfg.Database.PoolLifetime != 2*time.Minute {
		t.Fatalf("typed values not parsed: HTTP=%#v DB=%#v", cfg.HTTP, cfg.Database)
	}
}

func TestLoadProcessEnvironmentOverridesDotenvFile(t *testing.T) {
	setValidEnvironment(t)
	envFile := filepath.Join(t.TempDir(), "service.env")
	if err := os.WriteFile(envFile, []byte("APP_NAME=from-file\nAPP_PORT=3001\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	t.Setenv("ENV_FILE", envFile)
	t.Setenv("APP_NAME", "from-process")
	t.Setenv("APP_PORT", "4100")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Name != "from-process" || cfg.HTTP.Port != 4100 {
		t.Fatalf("process environment did not win: app=%#v http=%#v", cfg.App, cfg.HTTP)
	}
}

func TestLoadReportsAllMissingRequiredValues(t *testing.T) {
	t.Setenv("ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))
	for _, key := range []string{"APP_NAME", "APP_URL", "AUTH_SERVICE_GRPC_TARGET", "SERVICE_ID", "INTERNAL_SECRET", "DB_HOST", "DB_USERNAME", "DB_NAME"} {
		t.Setenv(key, "")
	}
	t.Setenv("APP_ENV", "invalid")
	t.Setenv("APP_PORT", "0")
	t.Setenv("DB_PORT", "0")
	t.Setenv("DB_POOL_MAX", "0")
	t.Setenv("DB_POOL_LIFETIME", "0s")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
	for _, expected := range []string{"APP_NAME is required", "APP_URL is required", "APP_PORT", "AUTH_SERVICE_GRPC_TARGET", "DB_HOST", "DB_USERNAME", "DB_NAME"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("error %q does not contain %q", err, expected)
		}
	}
}

func TestHumaConfigUsesBuildInfoAndEnvironmentURL(t *testing.T) {
	original := buildinfo.Version
	buildinfo.Version = "v9.8.7"
	t.Cleanup(func() { buildinfo.Version = original })
	cfg := &Config{App: AppConfig{Name: "template-api", Env: "test", URL: "https://api.example.com"}, Docs: DocsConfig{Enabled: true}}

	humaCfg := cfg.HumaConfig()
	if humaCfg.Info.Version != "v9.8.7" {
		t.Fatalf("version = %q, want buildinfo version", humaCfg.Info.Version)
	}
	if humaCfg.DocsPath != "/docs" || humaCfg.OpenAPIPath != "/openapi.json" {
		t.Fatalf("unexpected docs paths: docs=%q openapi=%q", humaCfg.DocsPath, humaCfg.OpenAPIPath)
	}
	if len(humaCfg.Servers) != 1 || humaCfg.Servers[0].URL != "https://api.example.com" {
		t.Fatalf("unexpected servers: %#v", humaCfg.Servers)
	}
}

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))
	values := map[string]string{
		"APP_NAME": "template-api", "APP_ENV": "test", "APP_URL": "https://api.example.com", "APP_PORT": "3000",
		"AUTH_ENABLED": "true", "AUTH_SERVICE_GRPC_TARGET": "127.0.0.1:50051", "SERVICE_ID": "template", "INTERNAL_SECRET": "secret",
		"DB_HOST": "127.0.0.1", "DB_PORT": "5432", "DB_USERNAME": "postgres", "DB_NAME": "template", "DB_POOL_IDLE": "1", "DB_POOL_MAX": "2", "DB_POOL_LIFETIME": "1m",
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
}
