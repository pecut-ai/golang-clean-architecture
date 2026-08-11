package model

type AddressResponse struct {
	ID         string `json:"id"`
	Street     string `json:"street"`
	City       string `json:"city"`
	Province   string `json:"province"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

type ListAddressRequest struct {
	UserId    string `json:"-" validate:"required"`
	ContactId string `json:"-" validate:"required,max=100,uuid"`
}

type CreateAddressRequest struct {
	UserId     string `json:"-" validate:"required"`
	ContactId  string `json:"-" validate:"required,max=100,uuid"`
	Street     string `json:"street" validate:"required,max=255"`
	City       string `json:"city" validate:"required,max=255"`
	Province   string `json:"province" validate:"required,max=255"`
	PostalCode string `json:"postal_code" validate:"required,max=10"`
	Country    string `json:"country" validate:"required,max=100"`
}

type UpdateAddressRequest struct {
	UserId     string `json:"-" validate:"required"`
	ContactId  string `json:"-" validate:"required,max=100,uuid"`
	ID         string `json:"-" validate:"required,max=100,uuid"`
	Street     string `json:"street" validate:"required,max=255"`
	City       string `json:"city" validate:"required,max=255"`
	Province   string `json:"province" validate:"required,max=255"`
	PostalCode string `json:"postal_code" validate:"required,max=10"`
	Country    string `json:"country" validate:"required,max=100"`
}

type GetAddressRequest struct {
	UserId    string `json:"-" validate:"required"`
	ContactId string `json:"-" validate:"required,max=100,uuid"`
	ID        string `json:"-" validate:"required,max=100,uuid"`
}

type DeleteAddressRequest struct {
	UserId    string `json:"-" validate:"required"`
	ContactId string `json:"-" validate:"required,max=100,uuid"`
	ID        string `json:"-" validate:"required,max=100,uuid"`
}
