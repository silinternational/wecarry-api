package models

import (
	"errors"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
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
func (o *OrganizationDomain) Organization(tx *pop.Connection) (Organization, error) {
	if o.OrganizationID <= 0 {
		return Organization{}, errors.New("OrganizationID is not valid")
	}
	var organization Organization
	if err := tx.Find(&organization, o.OrganizationID); err != nil {
		return Organization{}, err
	}
	return organization, nil
}

// Create stores the OrganizationDomain data as a new record in the database.
func (o *OrganizationDomain) Create(tx *pop.Connection) error {
	return create(tx, o)
}

// FindByDomain finds a record by the domain name
func (o *OrganizationDomain) FindByDomain(tx *pop.Connection, domainName string) error {
	return tx.Where("domain = ?", domainName).First(o)
}

// Save wrap tx.Save() call to check for errors and operate on attached object
func (o *OrganizationDomain) Save(tx *pop.Connection) error {
	return save(tx, o)
}
