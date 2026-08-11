package route

import (
	"testing"

	"golang-clean-architecture/internal/config"

	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
)

func TestRegisterBuildsRuntimeOpenAPIContract(t *testing.T) {
	app := fiber.New()
	cfg := (&config.Config{App: config.AppConfig{Name: "test", Env: "test", URL: "https://api.example.com"}}).HumaConfig()
	api := humafiber.New(app, cfg)
	Register(api, Dependencies{})

	spec := api.OpenAPI()
	if spec.Paths["/api/me"] == nil || spec.Paths["/api/me"].Get == nil {
		t.Fatal("GET /api/me is missing from OpenAPI")
	}
	if spec.Paths["/api/users/_login"] != nil {
		t.Fatal("stale local login operation is still present")
	}
	operation := spec.Paths["/api/contacts"].Get
	if operation == nil || len(operation.Security) != 1 {
		t.Fatal("contact operation is missing bearer security")
	}
}
