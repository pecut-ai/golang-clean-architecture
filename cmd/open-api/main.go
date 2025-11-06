package main

import (
	"encoding/json"
	"fmt"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/delivery/http/route"
	"log"
	"os"
	"path/filepath"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config := config.NewViper()

	// Create a minimal Fiber app (won't be started)
	app := fiber.New()
	appName := config.GetString("APP_NAME")
	webPort := config.GetString("WEB_PORT")

	// Create Huma API with the same config as the main app
	humaConfig := huma.DefaultConfig(appName, "1.0.0")
	humaConfig.Servers = []*huma.Server{
		{URL: fmt.Sprintf("http://localhost:%s", webPort), Description: "Development server"},
	}

	// Add security scheme for basic auth
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"basic": {
			Type:        "http",
			Scheme:      "basic",
			Description: "Basic authentication",
		},
	}

	api := humafiber.New(app, humaConfig)

	route.RegisterHumaOperations(api)

	// Get the OpenAPI spec
	spec := api.OpenAPI()

	// Marshal to JSON
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal OpenAPI spec:", err)
	}

	// Determine output file
	outputFile := "api/openapi.json"
	if len(os.Args) > 1 {
		outputFile = os.Args[1]
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Write to file
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		log.Fatal("Failed to write OpenAPI spec:", err)
	}

	log.Printf("✅ OpenAPI spec generated successfully: %s\n", outputFile)
	log.Printf("📊 Total operations: %d\n", countOperations(spec))
	log.Printf("🏷️  Tags: %v\n", getTags(spec))
}

func countOperations(spec *huma.OpenAPI) int {
	count := 0
	for _, pathItem := range spec.Paths {
		if pathItem.Get != nil {
			count++
		}
		if pathItem.Post != nil {
			count++
		}
		if pathItem.Put != nil {
			count++
		}
		if pathItem.Patch != nil {
			count++
		}
		if pathItem.Delete != nil {
			count++
		}
	}
	return count
}

func getTags(spec *huma.OpenAPI) []string {
	tagMap := make(map[string]bool)
	for _, pathItem := range spec.Paths {
		operations := []*huma.Operation{
			pathItem.Get, pathItem.Post, pathItem.Put,
			pathItem.Patch, pathItem.Delete,
		}
		for _, op := range operations {
			if op != nil {
				for _, tag := range op.Tags {
					tagMap[tag] = true
				}
			}
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	return tags
}
