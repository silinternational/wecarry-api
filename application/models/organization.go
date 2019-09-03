package models

import (
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
	ID                  int                  `json:"id" db:"id"`
	CreatedAt           time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at" db:"updated_at"`
	Name                string               `json:"name" db:"name"`
	Url                 nulls.String         `json:"url" db:"url"`
	AuthType            string               `json:"auth_type" db:"auth_type"`
	AuthConfig          string               `json:"auth_config" db:"auth_config"`
	Uuid                uuid.UUID            `json:"uuid" db:"uuid"`
	Users               Users                `many_to_many:"user_organizations"`
	OrganizationDomains []OrganizationDomain `has_many:"organization_domains"`
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

func (o *Organization) FindByUUID(uuid string) error {

	if uuid == "" {
		return fmt.Errorf("error: access token must not be blank")
	}

	if err := DB.Where("uuid = ?", uuid).First(o); err != nil {
		return fmt.Errorf("error finding org by uuid: %s", err.Error())
	}

	return nil
}

func (o *Organization) FindByDomain(domain string) error {
	var orgDomain OrganizationDomain
	if err := DB.Where("domain = ?", domain).First(&orgDomain); err != nil {
		return fmt.Errorf("error finding organization_domain by domain: %s", err.Error())
	}

	if err := DB.Eager().Where("id = ?", orgDomain.OrganizationID).First(o); err != nil {
		return fmt.Errorf("error finding organization by domain: %s", err.Error())
	}

	return nil
}

func (o *Organization) AddDomain(domain string) (OrganizationDomain, error) {
	// make sure domain is not registered to another org first
	var orgDomain OrganizationDomain

	count, err := DB.Where("domain = ?", domain).Count(&orgDomain)
	if err != nil {
		return OrganizationDomain{}, err
	}

	if count > 0 {
		return OrganizationDomain{}, fmt.Errorf("this domain (%s) is already in use", domain)
	}

	orgDomain.Domain = domain
	orgDomain.OrganizationID = o.ID
	err = DB.Save(&orgDomain)
	if err != nil {
		return OrganizationDomain{}, err
	}

	return orgDomain, nil
}

func (o *Organization) RemoveDomain(domain string) error {
	// make sure domain belongs to org for removal
	for _, od := range o.OrganizationDomains {
		if od.Domain == domain {
			err := DB.Destroy(&od)
			if err != nil {
				return err
			}
			return o.loadDomains()
		}
	}

	return fmt.Errorf("domain %s not found for organization", domain)
}

func (o *Organization) loadDomains() error {
	var ods []OrganizationDomain
	err := DB.Where("organization_id = ?", o.ID).All(&ods)
	if err != nil {
		return err
	}

	o.OrganizationDomains = ods
	return nil
}
