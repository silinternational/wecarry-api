package models

import (
	"errors"
	"time"
)

type OrganizationDomain struct {
	ID             int          `json:"id" db:"id"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	OrganizationID int          `json:"organization_id" db:"organization_id"`
	Domain         string       `json:"domain" db:"domain"`
	Organization   Organization `belongs_to:"organizations"`
}

type OrganizationDomains []OrganizationDomain

// GetOrganizationUUID loads the Organization record and converts its UUID to its string representation.
func (o *OrganizationDomain) GetOrganizationUUID() (string, error) {
	if o.OrganizationID <= 0 {
		return "", errors.New("OrganizationID is not valid")
	}
	if err := DB.Load(o, "Organization"); err != nil {
		return "", err
	}
	return o.Organization.UUID.String(), nil
}
