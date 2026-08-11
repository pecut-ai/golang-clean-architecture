package dto

import "golang-clean-architecture/internal/model"

type ListContactsInput struct {
	Name  string `query:"name" doc:"Filter by first or last name" maxLength:"100"`
	Email string `query:"email" doc:"Filter by email" maxLength:"200"`
	Phone string `query:"phone" doc:"Filter by phone" maxLength:"20"`
	Page  int    `query:"page" doc:"One-based page number" default:"1" minimum:"1"`
	Size  int    `query:"size" doc:"Items per page" default:"10" minimum:"1" maximum:"100"`
}

func (i *ListContactsInput) Request(userID string) *model.SearchContactRequest {
	return &model.SearchContactRequest{UserId: userID, Name: i.Name, Email: i.Email, Phone: i.Phone, Page: i.Page, Size: i.Size}
}

type ContactBody struct {
	FirstName string `json:"first_name" minLength:"1" maxLength:"100"`
	LastName  string `json:"last_name,omitempty" required:"false" maxLength:"100"`
	Email     string `json:"email,omitempty" required:"false" format:"email" maxLength:"200"`
	Phone     string `json:"phone,omitempty" required:"false" maxLength:"20"`
}

type CreateContactInput struct{ Body ContactBody }

func (i *CreateContactInput) Request(userID string) *model.CreateContactRequest {
	return &model.CreateContactRequest{UserId: userID, FirstName: i.Body.FirstName, LastName: i.Body.LastName, Email: i.Body.Email, Phone: i.Body.Phone}
}

type ContactPathInput struct {
	ContactID string `path:"contactId" doc:"Contact UUID" format:"uuid"`
}

func (i *ContactPathInput) GetRequest(userID string) *model.GetContactRequest {
	return &model.GetContactRequest{UserId: userID, ID: i.ContactID}
}

func (i *ContactPathInput) DeleteRequest(userID string) *model.DeleteContactRequest {
	return &model.DeleteContactRequest{UserId: userID, ID: i.ContactID}
}

func (i *ContactPathInput) ListAddressesRequest(userID string) *model.ListAddressRequest {
	return &model.ListAddressRequest{UserId: userID, ContactId: i.ContactID}
}

type UpdateContactInput struct {
	ContactID string `path:"contactId" doc:"Contact UUID" format:"uuid"`
	Body      ContactBody
}

func (i *UpdateContactInput) Request(userID string) *model.UpdateContactRequest {
	return &model.UpdateContactRequest{UserId: userID, ID: i.ContactID, FirstName: i.Body.FirstName, LastName: i.Body.LastName, Email: i.Body.Email, Phone: i.Body.Phone}
}

type AddressBody struct {
	Street     string `json:"street" minLength:"1" maxLength:"255"`
	City       string `json:"city" minLength:"1" maxLength:"255"`
	Province   string `json:"province" minLength:"1" maxLength:"255"`
	PostalCode string `json:"postal_code" minLength:"1" maxLength:"10"`
	Country    string `json:"country" minLength:"1" maxLength:"100"`
}

type CreateAddressInput struct {
	ContactID string `path:"contactId" doc:"Contact UUID" format:"uuid"`
	Body      AddressBody
}

func (i *CreateAddressInput) Request(userID string) *model.CreateAddressRequest {
	return &model.CreateAddressRequest{UserId: userID, ContactId: i.ContactID, Street: i.Body.Street, City: i.Body.City, Province: i.Body.Province, PostalCode: i.Body.PostalCode, Country: i.Body.Country}
}

type AddressPathInput struct {
	ContactID string `path:"contactId" doc:"Contact UUID" format:"uuid"`
	AddressID string `path:"addressId" doc:"Address UUID" format:"uuid"`
}

func (i *AddressPathInput) GetRequest(userID string) *model.GetAddressRequest {
	return &model.GetAddressRequest{UserId: userID, ContactId: i.ContactID, ID: i.AddressID}
}

func (i *AddressPathInput) DeleteRequest(userID string) *model.DeleteAddressRequest {
	return &model.DeleteAddressRequest{UserId: userID, ContactId: i.ContactID, ID: i.AddressID}
}

type UpdateAddressInput struct {
	ContactID string `path:"contactId" doc:"Contact UUID" format:"uuid"`
	AddressID string `path:"addressId" doc:"Address UUID" format:"uuid"`
	Body      AddressBody
}

func (i *UpdateAddressInput) Request(userID string) *model.UpdateAddressRequest {
	return &model.UpdateAddressRequest{UserId: userID, ContactId: i.ContactID, ID: i.AddressID, Street: i.Body.Street, City: i.Body.City, Province: i.Body.Province, PostalCode: i.Body.PostalCode, Country: i.Body.Country}
}
