package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/silinternational/handcarry-api/auth"
	"github.com/silinternational/handcarry-api/auth/saml"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

const AuthTypeSaml = "saml"

type Organization struct {
	ID                  int                   `json:"id" db:"id"`
	CreatedAt           time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at" db:"updated_at"`
	Name                string                `json:"name" db:"name"`
	Url                 nulls.String          `json:"url" db:"url"`
	AuthType            string                `json:"auth_type" db:"auth_type"`
	AuthConfig          string                `json:"auth_config" db:"auth_config"`
	Uuid                uuid.UUID             `json:"uuid" db:"uuid"`
	Users               Users                 `many_to_many:"user_organizations"`
	OrganizationDomains []OrganizationDomains `has_many:"organization_domains"`
}

// String is not required by pop and may be deleted
func (o Organization) String() string {
	jo, _ := json.Marshal(o)
	return string(jo)
}

// Organizations is not required by pop and may be deleted
type Organizations []Organization

// String is not required by pop and may be deleted
func (o Organizations) String() string {
	jo, _ := json.Marshal(o)
	return string(jo)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (o *Organization) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: o.Name, Name: "Name"},
		&validators.StringIsPresent{Field: o.AuthType, Name: "AuthType"},
		&validators.UUIDIsPresent{Field: o.Uuid, Name: "Uuid"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (o *Organization) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (o *Organization) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func (o *Organization) GetAuthProvider() (auth.Provider, error) {
	if o.AuthType == AuthTypeSaml {
		return saml.New([]byte(o.AuthConfig))
	}

	return &auth.EmptyProvider{}, fmt.Errorf("unsupported auth provider type: %s", o.AuthType)
}

func FindOrgByUUID(uuid string) (Organization, error) {

	if uuid == "" {
		return Organization{}, fmt.Errorf("error: access token must not be blank")
	}

	org := Organization{}

	queryString := fmt.Sprintf("uuid = '%s'", uuid)

	if err := DB.Where(queryString).First(&org); err != nil {
		return Organization{}, fmt.Errorf("error finding org by uuid: %s", err.Error())
	}

	return org, nil
}

func OrganizationFindByDomain(domain string) (Organization, error) {
	var orgDomain OrganizationDomains
	if err := DB.Eager().Where("domain = ?", domain).First(&orgDomain); err != nil {
		return Organization{}, fmt.Errorf("error finding organization by domain: %s", err.Error())
	}

	return orgDomain.Organization, nil
}
