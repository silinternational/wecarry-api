package api

import (
	"github.com/gofrs/uuid"
)

// swagger:model
type Watches []Watch

// A Watch for specified filter criteria.
// New requests matching all of the given criteria will generate a new notification.
// swagger:model
type Watch struct {
	// unique identifier for the Watch
	// swagger:strfmt uuid4
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Name is a short description, as named by the Watch creator
	Name string `json:"name"`

	// Destination to watch. If a new request has a destination near this location, a notification will be sent.
	Destination *Location `json:"destination,omitempty"`

	// Origin to watch. If a new request has an origin near this location, a notification will be sent.
	Origin *Location `json:"origin,omitempty"`

	// Meeting to watch. Notifications will be sent for new requests tied to this event.
	Meeting *Meeting `json:"meeting,omitempty"`

	// Search by text in request `title` or `description`
	SearchText string `json:"search_text"`

	// Maximum size of a requested item
	Size string `json:"size"`
}