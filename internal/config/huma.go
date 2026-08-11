package config

import (
	"fmt"

	"golang-clean-architecture/internal/buildinfo"

	"github.com/danielgtaylor/huma/v2"
)

func (c *Config) HumaConfig() huma.Config {
	cfg := huma.DefaultConfig(c.App.Name, buildinfo.Version)
	cfg.Info.Title = fmt.Sprintf("%s API", c.App.Name)
	cfg.Info.Description = "HTTP API for " + c.App.Name + ". Authentication is delegated to auth-service."
	cfg.OpenAPIPath = "/openapi.json"
	if c.Docs.Enabled {
		cfg.DocsPath = "/docs"
	} else {
		cfg.DocsPath = ""
	}
	cfg.Servers = []*huma.Server{{URL: c.App.URL, Description: c.App.Env}}
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Access token issued by auth-service",
		},
	}
	return cfg
}
