package api

import "github.com/gofrs/uuid"

// swagger:model
type Requests []Request

// Request is a hand carry request
//
// swagger:model
type Request struct {
	// unique id (uuid) for thread
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"uuid"`

	// Description of this request
	//
	// read-only: true
	Description string `json:"description"`

	// User who created this request
	//
	// read-only: true
	CreatedBy *User `json:"created_by"`
}
