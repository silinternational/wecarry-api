package models

import (
	"errors"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type OrganizationDomain struct {
	ID             int       `json:"id" db:"id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	OrganizationID int       `json:"organization_id" db:"organization_id"`
	Domain         string    `json:"domain" db:"domain"`
	AuthType       AuthType  `json:"auth_type" db:"auth_type"`
	AuthConfig     string    `json:"auth_config" db:"auth_config"`
}

type OrganizationDomains []OrganizationDomain

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (o *OrganizationDomain) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: o.OrganizationID, Name: "OrganizationID"},
		&validators.StringIsPresent{Field: o.Domain, Name: "Domain"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (o *OrganizationDomain) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Organization loads the Organization record
func (o *OrganizationDomain) Organization() (Organization, error) {
	if o.OrganizationID <= 0 {
		return Organization{}, errors.New("OrganizationID is not valid")
	}
	var organization Organization
	if err := DB.Find(&organization, o.OrganizationID); err != nil {
		return Organization{}, err
	}
	return organization, nil
}

// Create stores the OrganizationDomain data as a new record in the database.
func (o *OrganizationDomain) Create() error {
	return create(o)
}

// FindByDomain finds a record by the domain name
func (o *OrganizationDomain) FindByDomain(domainName string) error {
	return DB.Where("domain = ?", domainName).First(o)
}

// Save wrap DB.Save() call to check for errors and operate on attached object
func (o *OrganizationDomain) Save() error {
	return save(o)
}
