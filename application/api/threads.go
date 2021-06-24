package api

import (
	"time"

	"github.com/gofrs/uuid"
)

// Threads is a list of conversations that each have a list of messages between users
//
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
	ID uuid.UUID `json:"id"`

	// LastViewedAt is the time the auth user last viewed this thread. Messages with `updatedAt` after this time can be considered unread.
	LastViewedAt *time.Time `json:"last_viewed_at"`

	// Messages in this thread
	Messages *Messages `json:"messages"`

	// Users participating in the message thread. The request creator is automatically added to all of the requests's threads
	Participants *Users `json:"participants"`

	// Request that owns this message thread
	Request *Request `json:"request"`

	// UnreadMessageCount is the number of messages unread by the auth user
	UnreadMessageCount int `json:"unread_message_count"`

	// UpdatedAt = the time this thread was last updated or messages added to the thread
	UpdatedAt time.Time
}

// MarkMessagesAsReadInput is an object for setting the last_viewed_at time of a thread
// swagger:model
type MarkMessagesAsReadInput struct {
	// unique id (uuid) for thread
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ThreadID string `json:"threadID"`

	Time time.Time `json:"time"`
}
