package api

import "github.com/gofrs/uuid"

// Organization subscribed to the App. Provides privacy controls for visibility of Requests and Meetings, and specifies
// authentication for associated users.
// swagger:model
type Organization struct {
	// unique identifier for the Organization
	// swagger:strfmt uuid4
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Organization name, limited to 255 characters
	Name string `json:"name"`
}
