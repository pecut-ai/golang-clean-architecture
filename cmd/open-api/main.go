package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang-clean-architecture/internal/config"
	httproute "golang-clean-architecture/internal/delivery/http/route"

	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal("load configuration", err)
	}
	app := fiber.New()
	api := humafiber.New(app, cfg.HumaConfig())
	httproute.Register(api, httproute.Dependencies{})

	data, err := json.MarshalIndent(api.OpenAPI(), "", "  ")
	if err != nil {
		fatal("marshal OpenAPI", err)
	}
	output := "api/openapi.json"
	if len(os.Args) > 1 {
		output = os.Args[1]
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		fatal("create output directory", err)
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		fatal("write OpenAPI document", err)
	}
	fmt.Printf("OpenAPI document written to %s\n", output)
}

func fatal(action string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", action, err)
	os.Exit(1)
}
