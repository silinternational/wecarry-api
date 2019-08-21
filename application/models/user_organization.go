package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

const UserOrganizationRoleMember = "member"
const UserOrganizationRoleAdmin = "admin"

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
		&validators.StringIsPresent{Field: u.AuthID, Name: "AuthID"},
		&validators.EmailIsPresent{Field: u.AuthEmail, Name: "AuthEmail"},
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

// UserOrganizationFindByAuthEmail finds and returns an array of user organizations for the given email address
func UserOrganizationFindByAuthEmail(authEmail string, orgID int) ([]UserOrganization, error) {
	// Validate email address before query
	errs := validate.Validate(&validators.EmailIsPresent{Field: authEmail})
	if len(errs.Errors) > 0 {
		return []UserOrganization{}, fmt.Errorf("email address provided (%s) is not valid", authEmail)
	}

	where := "auth_email = ?"
	params := []interface{}{authEmail}

	if orgID != 0 {
		where += " AND organization_id = ?"
		params = append(params, orgID)
	}

	var userOrgs []UserOrganization
	if err := DB.Eager().Where(where, params...).All(&userOrgs); err != nil {
		return []UserOrganization{}, fmt.Errorf("error finding user by email: %s", err.Error())
	}

	return userOrgs, nil
}

func FindUserOrganization(user User, org Organization) (UserOrganization, error) {
	var userOrg UserOrganization
	if err := DB.Where("user_id = ? AND organization_id = ?", user.ID, org.ID).First(&userOrg); err != nil {
		return UserOrganization{}, fmt.Errorf("association not found for user '%v' and org '%v' (%s)", user.Nickname, org.Name, err.Error())
	}

	return userOrg, nil
}
