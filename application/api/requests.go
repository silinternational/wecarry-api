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
// TODO: reconcile api.Request struct with the UI field list
type Request struct {
	// unique identifier for the Request
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Whether request is editable by current user
	IsEditable bool `json:"isEditable"`

	// Request status: OPEN, ACCEPTED, DELIVERED, RECEIVED, COMPLETED, REMOVED
	RequestStatus string `json:"status"`

	// Profile of the user that created this request
	CreatedBy *User `json:"created_by"`

	// Profile of the user that is the provider for this request
	Provider *User `json:"provider,omitempty"`

	// Users that have offered to carry this request
	PotentialProviders *[]User `json:"potentialProviders"`

	// Organization associated with this request
	Organization *Organization `json:"organization,omitempty"`

	// Short description of item, limited to 255 characters
	Title string `json:"title"`

	// Optional, longer description of the item, limited to 4,096 characters
	Description string `json:"description"`

	// Geographic location where item is needed
	Destination *Location `json:"destination"`

	// Optional geographic location where the item can be picked up, purchased, or otherwise obtained
	Origin *Location `json:"origin,omitempty"`

	// Broad category of the size of item
	Size string `json:"size"`

	// List of message threads associated with this request
	Threads *Threads `json:"threads"`

	// Date and time this request was created
	CreatedAt *time.Time `json:"createdAt"`

	// Date and time this request was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// Date (yyyy-mm-dd) before which the item will be needed. The record may be hidden or removed after this date
	NeededBefore time.Time `json:"neededBefore"`

	// Optional weight of the item, measured in kilograms
	Kilograms float64 `json:"kilograms"`

	// Optional URL to further describe or point to detail about the item, limited to 255 characters
	Url string `json:"url"`

	// Photo of the item
	Photo *File `json:"photo"`
}

// swagger:model
type RequestsAbridged []RequestAbridged

// Request is a hand carry request
//
// swagger:model
type RequestAbridged struct {
	// unique identifier for the Request
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Profile of the user that created this request
	CreatedBy *User `json:"created_by"`

	// Profile of the user that is the provider for this request
	Provider *User `json:"provider,omitempty"`

	// Short description of item, limited to 255 characters
	Title string `json:"title"`

	// Geographic location where item is needed
	Destination *Location `json:"destination"`

	// Optional geographic location where the item can be picked up, purchased, or otherwise obtained
	Origin *Location `json:"origin,omitempty"`

	// Broad category of the size of item
	Size string `json:"size"`

	// Photo of the item
	Photo *File `json:"photo"`
}
