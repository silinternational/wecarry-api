package api

import (
	"time"

	"github.com/gofrs/uuid"

)

// swagger:model
type Requests []Request

// Request is a hand carry request
//
// swagger:model
type Request struct {
	// unique id (uuid) for thread
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Whether request is editable by current user
	IsEditable bool `json:"isEditable"`

	// Request status: OPEN, ACCEPTED, DELIVERED, RECEIVED, COMPLETED, REMOVED
	RequestStatus string `json:"status"`

	// Profile of the user that created this request.
	CreatedBy *User `json:"created_by"`

	// Request provider
	Provider *User `json:"provider,omitempty"`

	// Potential providers
	PotentialProviders *[]User `json:"potentialProviders"`//,omitempty"`

	// Organization associated with request
	Organization *Organization `json:"organization,omitempty"`

	// Short request description
	Title string `json:"title"`

	// Long request description
	Description string `json:"description"`

	Destination *Location `json:"destination"`

	Origin *Location `json:"origin,omitempty"`

	Size string `json:"size"`

	Threads *Threads `json:"threads"`//,omitempty"`

	CreatedAt *time.Time `json:"createdAt"`

	UpdatedAt time.Time `json:"updatedAt"`

	Kilograms float64 `json:"kilograms"`

	Url string `json:"url"`

}
