package api

import (
	"github.com/gobuffalo/nulls"
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
	SearchText nulls.String `json:"search_text"`

	// Maximum size of a requested item
	Size nulls.String `json:"size"`
}

// Input object to create a new Watch for the user
// swagger:model
type WatchInput struct {
	// Short description, as named by the Watch creator
	Name string `json:"name"`

	// Destination to watch. If a new request has a destination near this location, a notification will be sent.
	Destination *Location `json:"destination"`

	// Origin to watch. If a new request has an origin near this location, a notification will be sent.
	Origin *Location `json:"origin"`

	// Meeting to watch. Notifications will be sent for new requests tied to this event.
	MeetingID *string `json:"meeting_id"`

	// Search by text in `title` or `description`
	SearchText *string `json:"search_text"`

	// Maximum size of a requested item
	Size *RequestSize `json:"size"`
}

func (w WatchInput) IsEmpty() bool {
	if w.Destination != nil || w.Origin != nil {
		return false
	}

	if w.SearchText != nil && *w.SearchText != "" {
		return false
	}

	if w.Size != nil && w.Size.String() != "" {
		return false
	}

	if w.MeetingID != nil && *w.MeetingID != "" {
		return false
	}

	return true
}
