package api

import (
	"time"

	"github.com/gofrs/uuid"
)

// swagger:model
type Messages []Message

// A Message is one element associated with a conversation thread
//
// swagger:model
type Message struct {
	// unique id (uuid) for message
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// datetime when message was created
	//
	// read-only: true
	// example: 2020-10-02T15:00:00Z
	CreatedAt time.Time `json:"created_at"`

	// Text content of the message
	//
	Content string `json:"content"`

	// User who sent the message
	//
	// read-only: true
	SentBy *User `json:"sender"`
}

// MessageInput contains parameters to create a Message
// swagger:model
type MessageInput struct {
	// message content, limited to 4,096 characters
	Content string `json:"content"`

	// ID of the subject Request
	RequestID string `json:"request_id"`

	// Message thread to which the new message should be attached. If not specified, a new thread is created.
	ThreadID *string `json:"thread_id"`
}
