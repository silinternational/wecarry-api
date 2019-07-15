package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type User struct {
	ID           int               `json:"id" db:"id"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
	Email        string            `json:"email" db:"email"`
	FirstName    string            `json:"first_name" db:"first_name"`
	LastName     string            `json:"last_name" db:"last_name"`
	Nickname     string            `json:"nickname" db:"nickname"`
	AuthOrgID    int               `json:"auth_org_id" db:"auth_org_id"`
	AuthOrgUid   string            `json:"auth_org_uid" db:"auth_org_uid"`
	AdminRole    nulls.String      `json:"admin_role" db:"admin_role"`
	Uuid         string            `json:"uuid" db:"uuid"`
	AuthOrg      Organization      `belongs_to:"organizations"`
	AccessTokens []UserAccessToken `has_many:"user_access_tokens"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: u.ID, Name: "ID"},
		&validators.StringIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.FirstName, Name: "FirstName"},
		&validators.StringIsPresent{Field: u.LastName, Name: "LastName"},
		&validators.StringIsPresent{Field: u.Nickname, Name: "Nickname"},
		&validators.IntIsPresent{Field: u.AuthOrgID, Name: "AuthOrgID"},
		&validators.StringIsPresent{Field: u.AuthOrgUid, Name: "AuthOrgUid"},
		&validators.StringIsPresent{Field: u.Uuid, Name: "Uuid"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
