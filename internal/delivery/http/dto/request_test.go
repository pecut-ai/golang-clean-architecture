package dto

import "testing"

func TestCreateContactInputMapsAuthenticatedUserOutsideBody(t *testing.T) {
	input := &CreateContactInput{Body: ContactBody{FirstName: "Ada", Email: "ada@example.com"}}

	request := input.Request("auth-user-123")

	if request.UserId != "auth-user-123" || request.FirstName != "Ada" || request.Email != "ada@example.com" {
		t.Fatalf("unexpected application request: %#v", request)
	}
}

func TestUpdateAddressInputMapsPathAndBody(t *testing.T) {
	input := &UpdateAddressInput{
		ContactID: "2a3492d8-ff88-4521-a820-71e858cadf58",
		AddressID: "bb58405f-4692-4520-a1d0-249288664f39",
		Body:      AddressBody{Street: "Main Street", City: "Jakarta", Province: "DKI Jakarta", PostalCode: "10110", Country: "Indonesia"},
	}

	request := input.Request("auth-user-123")

	if request.UserId != "auth-user-123" || request.ContactId != input.ContactID || request.ID != input.AddressID || request.Street != "Main Street" {
		t.Fatalf("unexpected application request: %#v", request)
	}
}
