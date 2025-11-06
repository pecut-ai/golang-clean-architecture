package route

import (
	"context"
	"golang-clean-architecture/internal/delivery/http"
	"golang-clean-architecture/internal/model"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App               *fiber.App
	Api               huma.API
	UserController    *http.UserController
	ContactController *http.ContactController
	AddressController *http.AddressController
	AuthMiddleware    fiber.Handler
}

func (c *RouteConfig) Setup() {
	c.SetupGuestRoute()
	c.SetupAuthRoute()
}

func (c *RouteConfig) SetupGuestRoute() {
	// Fiber routing
	if c.UserController != nil {
		c.App.Post("/api/users", c.UserController.Register)
		c.App.Post("/api/users/_login", c.UserController.Login)
	}
}

func (c *RouteConfig) SetupAuthRoute() {
	// Fiber routing with authentication middleware
	if c.AuthMiddleware != nil {
		c.App.Use(c.AuthMiddleware)
	}

	// User routes
	if c.UserController != nil {
		c.App.Delete("/api/users", c.UserController.Logout)
		c.App.Patch("/api/users/_current", c.UserController.Update)
		c.App.Get("/api/users/_current", c.UserController.Current)
	}

	// Contact routes
	if c.ContactController != nil {
		c.App.Get("/api/contacts", c.ContactController.List)
		c.App.Post("/api/contacts", c.ContactController.Create)
		c.App.Put("/api/contacts/:contactId", c.ContactController.Update)
		c.App.Get("/api/contacts/:contactId", c.ContactController.Get)
		c.App.Delete("/api/contacts/:contactId", c.ContactController.Delete)
	}

	// Address routes
	if c.AddressController != nil {
		c.App.Get("/api/contacts/:contactId/addresses", c.AddressController.List)
		c.App.Post("/api/contacts/:contactId/addresses", c.AddressController.Create)
		c.App.Put("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Update)
		c.App.Get("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Get)
		c.App.Delete("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Delete)
	}
}

func RegisterHumaOperations(api huma.API) {
	registerGuestHumaOperations(api)
	registerAuthHumaOperations(api)
}

// registerGuestHumaOperations registers Huma operations for public routes
func registerGuestHumaOperations(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "register-user",
		Method:      "POST",
		Path:        "/api/users",
		Summary:     "Register a new user",
		Description: "Create a new user account with username and password",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *struct {
		Body model.RegisterUserRequest
	}) (*struct {
		Body model.WebResponse[*model.UserResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "login-user",
		Method:      "POST",
		Path:        "/api/users/_login",
		Summary:     "Login user",
		Description: "Authenticate user and receive a bearer token",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *struct {
		Body model.LoginUserRequest
	}) (*struct {
		Body model.WebResponse[*model.UserResponse]
	}, error) {
		return nil, nil
	})
}

// registerAuthHumaOperations registers Huma operations for authenticated routes
func registerAuthHumaOperations(api huma.API) {
	// User operations
	huma.Register(api, huma.Operation{
		OperationID: "logout-user",
		Method:      "DELETE",
		Path:        "/api/users",
		Summary:     "Logout current user",
		Description: "Invalidate the current user's authentication token",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct{}) (*struct{ Body model.WebResponse[bool] }, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-current-user",
		Method:      "PATCH",
		Path:        "/api/users/_current",
		Summary:     "Update current user",
		Description: "Update the authenticated user's profile information",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		Body model.UpdateUserRequest
	}) (*struct {
		Body model.WebResponse[*model.UserResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-current-user",
		Method:      "GET",
		Path:        "/api/users/_current",
		Summary:     "Get current user",
		Description: "Retrieve the authenticated user's profile information",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct{}) (*struct {
		Body model.WebResponse[*model.UserResponse]
	}, error) {
		return nil, nil
	})

	// Contact operations
	huma.Register(api, huma.Operation{
		OperationID: "list-contacts",
		Method:      "GET",
		Path:        "/api/contacts",
		Summary:     "List contacts with search",
		Description: "Search and filter contacts with pagination support",
		Tags:        []string{"Contacts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		Name  string `query:"name" doc:"Filter by name"`
		Email string `query:"email" doc:"Filter by email"`
		Phone string `query:"phone" doc:"Filter by phone"`
		Page  int    `query:"page" doc:"Page number" default:"1"`
		Size  int    `query:"size" doc:"Page size" default:"10"`
	}) (*struct {
		Body model.WebResponse[[]model.ContactResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-contact",
		Method:      "POST",
		Path:        "/api/contacts",
		Summary:     "Create a new contact",
		Description: "Add a new contact to the authenticated user's contact list",
		Tags:        []string{"Contacts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		Body model.CreateContactRequest
	}) (*struct {
		Body model.WebResponse[*model.ContactResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-contact",
		Method:      "PUT",
		Path:        "/api/contacts/{contactId}",
		Summary:     "Update a contact",
		Description: "Update an existing contact's information",
		Tags:        []string{"Contacts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		ContactID string `path:"contactId" doc:"Contact ID"`
		Body      model.UpdateContactRequest
	}) (*struct {
		Body model.WebResponse[*model.ContactResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-contact",
		Method:      "GET",
		Path:        "/api/contacts/{contactId}",
		Summary:     "Get a contact by ID",
		Description: "Retrieve detailed information about a specific contact",
		Tags:        []string{"Contacts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		ContactID string `path:"contactId" doc:"Contact ID"`
	}) (*struct {
		Body model.WebResponse[*model.ContactResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-contact",
		Method:      "DELETE",
		Path:        "/api/contacts/{contactId}",
		Summary:     "Delete a contact",
		Description: "Remove a contact from the authenticated user's contact list",
		Tags:        []string{"Contacts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		ContactID string `path:"contactId" doc:"Contact ID"`
	}) (*struct{ Body model.WebResponse[bool] }, error) {
		return nil, nil
	})

	// Address operations
	huma.Register(api, huma.Operation{
		OperationID: "list-addresses",
		Method:      "GET",
		Path:        "/api/contacts/{contactId}/addresses",
		Summary:     "List addresses for a contact",
		Description: "Retrieve all addresses associated with a specific contact",
		Tags:        []string{"Addresses"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		ContactID string `path:"contactId" doc:"Contact ID"`
	}) (*struct {
		Body model.WebResponse[[]model.AddressResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-address",
		Method:      "POST",
		Path:        "/api/contacts/{contactId}/addresses",
		Summary:     "Create a new address",
		Description: "Add a new address to a contact",
		Tags:        []string{"Addresses"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		ContactID string `path:"contactId" doc:"Contact ID"`
		Body      model.CreateAddressRequest
	}) (*struct {
		Body model.WebResponse[*model.AddressResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-address",
		Method:      "PUT",
		Path:        "/api/contacts/{contactId}/addresses/{addressId}",
		Summary:     "Update an address",
		Description: "Update an existing address for a contact",
		Tags:        []string{"Addresses"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		ContactID string `path:"contactId" doc:"Contact ID"`
		AddressID string `path:"addressId" doc:"Address ID"`
		Body      model.UpdateAddressRequest
	}) (*struct {
		Body model.WebResponse[*model.AddressResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-address",
		Method:      "GET",
		Path:        "/api/contacts/{contactId}/addresses/{addressId}",
		Summary:     "Get an address by ID",
		Description: "Retrieve detailed information about a specific address",
		Tags:        []string{"Addresses"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		ContactID string `path:"contactId" doc:"Contact ID"`
		AddressID string `path:"addressId" doc:"Address ID"`
	}) (*struct {
		Body model.WebResponse[*model.AddressResponse]
	}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-address",
		Method:      "DELETE",
		Path:        "/api/contacts/{contactId}/addresses/{addressId}",
		Summary:     "Delete an address",
		Description: "Remove an address from a contact",
		Tags:        []string{"Addresses"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		ContactID string `path:"contactId" doc:"Contact ID"`
		AddressID string `path:"addressId" doc:"Address ID"`
	}) (*struct{ Body model.WebResponse[bool] }, error) {
		return nil, nil
	})
}
