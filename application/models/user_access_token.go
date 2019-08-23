package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type UserAccessToken struct {
	ID                 int              `json:"id" db:"id"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
	UserID             int              `json:"user_id" db:"user_id"`
	UserOrganizationID int              `json:"user_organization_id" db:"user_organization_id"`
	AccessToken        string           `json:"access_token" db:"access_token"`
	ExpiresAt          time.Time        `json:"expires_at" db:"expires_at"`
	User               User             `belongs_to:"users"`
	UserOrganization   UserOrganization `belongs_to:"user_organizations"`
}

// String is not required by pop and may be deleted
func (u UserAccessToken) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// UserAccessTokens is not required by pop and may be deleted
type UserAccessTokens []UserAccessToken

// String is not required by pop and may be deleted
func (u UserAccessTokens) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *UserAccessToken) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: u.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: u.AccessToken, Name: "AccessToken"},
		&validators.TimeIsPresent{Field: u.ExpiresAt, Name: "ExpiresAt"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *UserAccessToken) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *UserAccessToken) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func DeleteAccessToken(accessToken string) error {
	userAccessToken := UserAccessToken{
		AccessToken: hashClientIdAccessToken(accessToken),
	}
	err := DB.Where("access_token = ?", userAccessToken.AccessToken).First(&userAccessToken)
	if err != nil {
		return err
	}

	return DB.Destroy(&userAccessToken)
}

func (u *UserAccessToken) FindByBearerToken(bearerToken string) error {
	if err := DB.Eager().Where("access_token = ?", hashClientIdAccessToken(bearerToken)).First(u); err != nil {
		return fmt.Errorf("failed to find access token, %v", err)
	}
	return nil
}
