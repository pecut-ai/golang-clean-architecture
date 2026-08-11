package dto

import "golang-clean-architecture/internal/model"

type Envelope[T any] struct {
	Data   T             `json:"data"`
	Paging *PageMetadata `json:"paging,omitempty"`
}

type PageMetadata struct {
	Page      int   `json:"page"`
	Size      int   `json:"size"`
	TotalItem int64 `json:"total_item"`
	TotalPage int64 `json:"total_page"`
}

type HealthResponse struct {
	Status  string `json:"status" doc:"Process health status" example:"ok"`
	Version string `json:"version" doc:"Build version" example:"dev"`
}

type HealthOutput struct{ Body HealthResponse }

type CurrentUserResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	RoleID      *int32   `json:"role_id,omitempty"`
	RoleName    *string  `json:"role_name,omitempty"`
	SessionID   string   `json:"session_id"`
	Permissions []string `json:"permissions"`
}

type CurrentUserOutput struct{ Body Envelope[CurrentUserResponse] }
type ContactOutput struct {
	Body Envelope[*model.ContactResponse]
}
type ContactsOutput struct {
	Body Envelope[[]model.ContactResponse]
}
type AddressOutput struct {
	Body Envelope[*model.AddressResponse]
}
type AddressesOutput struct {
	Body Envelope[[]model.AddressResponse]
}
type BooleanOutput struct{ Body Envelope[bool] }
