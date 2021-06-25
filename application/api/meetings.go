package api

import (
	"time"

	"github.com/gofrs/uuid"
)

// Meeting a/k/a Event, to serve as a focal point for finding, answering, carrying, and exchanging requests
// swagger:model
type Meeting struct {
	// unique id (uuid) for thread
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Short name, limited to 80 characters
	Name string `json:"name"`

	// the text-only description, limited to 4096 characters"
	Description string `json:"description"`

	// The date of the first day of the meeting (event)
	StartDate time.Time `json:"start_date"`

	// The date of the last day of the meeting (event)
	EndDate time.Time `json:"end_date"`

	// The time the meeting (event) was added to the app"
	CreatedAt time.Time `json:"created_at"`

	// The time the meeting (event) was last modified in the app"
	UpdatedAt time.Time `json:"updated_at"`

	// meeting (event) information URL -- should be a full website, but could be an information document such as a pdf"
	MoreInfoURL string `json:"more_info_url"`
}
