package route

import (
	"context"
	"errors"
	"math"

	"golang-clean-architecture/internal/buildinfo"
	"golang-clean-architecture/internal/delivery/http/dto"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/usecase"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
	authservice "github.com/pecut-ai/auth-service/pkg/v2/auth"
)

type Dependencies struct {
	Contact *usecase.ContactUseCase
	Address *usecase.AddressUseCase
}

type Handler struct{ deps Dependencies }

func Register(api huma.API, deps Dependencies) {
	h := &Handler{deps: deps}
	h.registerHealth(api)
	h.registerCurrentUser(api)
	h.registerContacts(api)
	h.registerAddresses(api)
}

func (h *Handler) registerHealth(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-health",
		Method:      fiber.MethodGet,
		Path:        "/api/health",
		Summary:     "Check process health",
		Tags:        []string{"System"},
	}, func(context.Context, *struct{}) (*dto.HealthOutput, error) {
		return &dto.HealthOutput{Body: dto.HealthResponse{Status: "ok", Version: buildinfo.Version}}, nil
	})
}

func (h *Handler) registerCurrentUser(api huma.API) {
	huma.Register(api, securedOperation("get-current-user", fiber.MethodGet, "/api/me", "Get authenticated user context", "Authentication"), h.getCurrentUser)
}

func (h *Handler) registerContacts(api huma.API) {
	huma.Register(api, securedOperation("list-contacts", fiber.MethodGet, "/api/contacts", "List contacts", "Contacts"), h.listContacts)
	huma.Register(api, securedOperation("create-contact", fiber.MethodPost, "/api/contacts", "Create a contact", "Contacts"), h.createContact)
	huma.Register(api, securedOperation("update-contact", fiber.MethodPut, "/api/contacts/{contactId}", "Update a contact", "Contacts"), h.updateContact)
	huma.Register(api, securedOperation("get-contact", fiber.MethodGet, "/api/contacts/{contactId}", "Get a contact", "Contacts"), h.getContact)
	huma.Register(api, securedOperation("delete-contact", fiber.MethodDelete, "/api/contacts/{contactId}", "Delete a contact", "Contacts"), h.deleteContact)
}

func (h *Handler) registerAddresses(api huma.API) {
	huma.Register(api, securedOperation("list-addresses", fiber.MethodGet, "/api/contacts/{contactId}/addresses", "List contact addresses", "Addresses"), h.listAddresses)
	huma.Register(api, securedOperation("create-address", fiber.MethodPost, "/api/contacts/{contactId}/addresses", "Create a contact address", "Addresses"), h.createAddress)
	huma.Register(api, securedOperation("update-address", fiber.MethodPut, "/api/contacts/{contactId}/addresses/{addressId}", "Update a contact address", "Addresses"), h.updateAddress)
	huma.Register(api, securedOperation("get-address", fiber.MethodGet, "/api/contacts/{contactId}/addresses/{addressId}", "Get a contact address", "Addresses"), h.getAddress)
	huma.Register(api, securedOperation("delete-address", fiber.MethodDelete, "/api/contacts/{contactId}/addresses/{addressId}", "Delete a contact address", "Addresses"), h.deleteAddress)
}

func securedOperation(id, method, path, summary, tag string) huma.Operation {
	return huma.Operation{OperationID: id, Method: method, Path: path, Summary: summary, Tags: []string{tag}, Security: []map[string][]string{{"bearerAuth": {}}}}
}

func (h *Handler) getCurrentUser(ctx context.Context, _ *struct{}) (*dto.CurrentUserOutput, error) {
	user, ok := authservice.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID == "" {
		return nil, huma.Error401Unauthorized("authenticated user is missing from request context")
	}
	return &dto.CurrentUserOutput{Body: dto.Envelope[dto.CurrentUserResponse]{Data: dto.CurrentUserResponse{
		ID: user.UserID, Username: user.Username, RoleID: user.RoleID, RoleName: user.RoleName, SessionID: user.SessionID, Permissions: user.Permissions,
	}}}, nil
}

func (h *Handler) listContacts(ctx context.Context, input *dto.ListContactsInput) (*dto.ContactsOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	contacts, total, err := h.deps.Contact.Search(ctx, input.Request(userID))
	if err != nil {
		return nil, httpError(err)
	}
	return &dto.ContactsOutput{Body: dto.Envelope[[]model.ContactResponse]{
		Data:   contacts,
		Paging: &dto.PageMetadata{Page: input.Page, Size: input.Size, TotalItem: total, TotalPage: int64(math.Ceil(float64(total) / float64(input.Size)))},
	}}, nil
}

func (h *Handler) createContact(ctx context.Context, input *dto.CreateContactInput) (*dto.ContactOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	contact, err := h.deps.Contact.Create(ctx, input.Request(userID))
	if err != nil {
		return nil, httpError(err)
	}
	return &dto.ContactOutput{Body: dto.Envelope[*model.ContactResponse]{Data: contact}}, nil
}

func (h *Handler) updateContact(ctx context.Context, input *dto.UpdateContactInput) (*dto.ContactOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	contact, err := h.deps.Contact.Update(ctx, input.Request(userID))
	if err != nil {
		return nil, httpError(err)
	}
	return &dto.ContactOutput{Body: dto.Envelope[*model.ContactResponse]{Data: contact}}, nil
}

func (h *Handler) getContact(ctx context.Context, input *dto.ContactPathInput) (*dto.ContactOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	contact, err := h.deps.Contact.Get(ctx, input.GetRequest(userID))
	if err != nil {
		return nil, httpError(err)
	}
	return &dto.ContactOutput{Body: dto.Envelope[*model.ContactResponse]{Data: contact}}, nil
}

func (h *Handler) deleteContact(ctx context.Context, input *dto.ContactPathInput) (*dto.BooleanOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.deps.Contact.Delete(ctx, input.DeleteRequest(userID)); err != nil {
		return nil, httpError(err)
	}
	return &dto.BooleanOutput{Body: dto.Envelope[bool]{Data: true}}, nil
}

func (h *Handler) listAddresses(ctx context.Context, input *dto.ContactPathInput) (*dto.AddressesOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	addresses, err := h.deps.Address.List(ctx, input.ListAddressesRequest(userID))
	if err != nil {
		return nil, httpError(err)
	}
	return &dto.AddressesOutput{Body: dto.Envelope[[]model.AddressResponse]{Data: addresses}}, nil
}

func (h *Handler) createAddress(ctx context.Context, input *dto.CreateAddressInput) (*dto.AddressOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	address, err := h.deps.Address.Create(ctx, input.Request(userID))
	if err != nil {
		return nil, httpError(err)
	}
	return &dto.AddressOutput{Body: dto.Envelope[*model.AddressResponse]{Data: address}}, nil
}

func (h *Handler) updateAddress(ctx context.Context, input *dto.UpdateAddressInput) (*dto.AddressOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	address, err := h.deps.Address.Update(ctx, input.Request(userID))
	if err != nil {
		return nil, httpError(err)
	}
	return &dto.AddressOutput{Body: dto.Envelope[*model.AddressResponse]{Data: address}}, nil
}

func (h *Handler) getAddress(ctx context.Context, input *dto.AddressPathInput) (*dto.AddressOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	address, err := h.deps.Address.Get(ctx, input.GetRequest(userID))
	if err != nil {
		return nil, httpError(err)
	}
	return &dto.AddressOutput{Body: dto.Envelope[*model.AddressResponse]{Data: address}}, nil
}

func (h *Handler) deleteAddress(ctx context.Context, input *dto.AddressPathInput) (*dto.BooleanOutput, error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.deps.Address.Delete(ctx, input.DeleteRequest(userID)); err != nil {
		return nil, httpError(err)
	}
	return &dto.BooleanOutput{Body: dto.Envelope[bool]{Data: true}}, nil
}

func authenticatedUserID(ctx context.Context) (string, error) {
	user, ok := authservice.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID == "" {
		return "", huma.Error401Unauthorized("authenticated user is missing from request context")
	}
	return user.UserID, nil
}

func httpError(err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return huma.NewError(fiberErr.Code, fiberErr.Message)
	}
	return err
}
