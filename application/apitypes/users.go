package apitypes

import "github.com/gofrs/uuid"

// swagger:model
type Users []User

// app user
// swagger:model
type User struct {
	// user ID
	//
	// read only: true
	// swagger:strfmt uuid4
	ID uuid.UUID `json:"uuid"`

	// user's nickname
	Nickname string `json:"nickname"`
}
