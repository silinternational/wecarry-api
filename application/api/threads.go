package api

import (
	"github.com/gofrs/uuid"
)

// swagger:model
type Threads []Thread

// A Thread is a list of messages in a conversation between users
//
// swagger:model
type Thread struct {
	// unique id (uuid) for thread
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"uuid"`

	// Messages in this thread
	//
	// read-only: true
	Messages *Messages `json:"messages"`

	// Users associated with this thread
	//
	// read-only: true
	Participants *Users `json:"participants"`

	// Request associated with this thread
	//
	// read-only: true
	Request *Request `json:"request"`
}
