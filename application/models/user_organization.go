package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type UserOrganization struct {
	ID             int          `json:"id" db:"id"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	OrganizationID int          `json:"organization_id" db:"organization_id"`
	UserID         int          `json:"user_id" db:"user_id"`
	Role           string       `json:"role" db:"role"`
	User           User         `belongs_to:"users"`
	Organization   Organization `belongs_to:"organizations"`
}

// String is not required by pop and may be deleted
func (u UserOrganization) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// UserOrganizations is not required by pop and may be deleted
type UserOrganizations []UserOrganization

// String is not required by pop and may be deleted
func (u UserOrganizations) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *UserOrganization) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: u.OrganizationID, Name: "OrganizationID"},
		&validators.IntIsPresent{Field: u.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: u.Role, Name: "Role"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *UserOrganization) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *UserOrganization) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
