package models

import "time"

type OrganizationDomain struct {
	ID             int          `json:"id" db:"id"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	OrganizationID int          `json:"organization_id" db:"organization_id"`
	Domain         string       `json:"domain" db:"domain"`
	Organization   Organization `belongs_to:"organizations"`
}
