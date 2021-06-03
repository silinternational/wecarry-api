package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
)

const (
	UserOrganizationRoleUser  = "user"
	UserOrganizationRoleAdmin = "admin"
)

type UserOrganization struct {
	ID             int          `json:"id" db:"id"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	OrganizationID int          `json:"organization_id" db:"organization_id"`
	UserID         int          `json:"user_id" db:"user_id"`
	Role           string       `json:"role" db:"role"`
	AuthID         string       `json:"auth_id" db:"auth_id"`
	AuthEmail      string       `json:"auth_email" db:"auth_email"`
	LastLogin      time.Time    `json:"last_login" db:"last_login"`
	User           User         `belongs_to:"users"`
	Organization   Organization `belongs_to:"organizations"`
}

// String can be helpful for serializing the model
func (u UserOrganization) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// UserOrganizations is merely for convenience and brevity
type UserOrganizations []UserOrganization

// String can be helpful for serializing the model
func (u UserOrganizations) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (u *UserOrganization) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: u.OrganizationID, Name: "OrganizationID"},
		&validators.IntIsPresent{Field: u.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: u.Role, Name: "Role"},
		&validators.StringIsPresent{Field: u.AuthID, Name: "AuthID"},
		&validators.EmailIsPresent{Field: u.AuthEmail, Name: "AuthEmail"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (u *UserOrganization) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (u *UserOrganization) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// FindByAuthEmail finds UserOrganizations for the given email address. However, if the
// orgID param is greater than zero, it will find only the one with both that authEmail and orgID.
func (u *UserOrganizations) FindByAuthEmail(authEmail string, orgID int) error {
	// Validate email address before query
	errs := validate.Validate(&validators.EmailIsPresent{Field: authEmail})
	if len(errs.Errors) > 0 {
		return fmt.Errorf("email address provided (%s) is not valid", authEmail)
	}

	where := "auth_email = ?"
	params := []interface{}{authEmail}

	if orgID > 0 {
		where += " AND organization_id = ?"
		params = append(params, orgID)
	}

	if err := DB.Eager().Where(where, params...).All(u); err != nil {
		return fmt.Errorf("error finding user by email: %s", err.Error())
	}

	return nil
}

// Create stores the UserOrganization data as a new record in the database.
func (u *UserOrganization) Create() error {
	return create(u)
}
