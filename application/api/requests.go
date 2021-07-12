package api

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

type (
	RequestStatus     string
	RequestVisibility string
)

// swagger:model
type Requests []Request

// Request is a hand carry request
//
// swagger:model
type Request struct {
	// unique identifier for the Request
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Whether request is editable by current user
	IsEditable bool `json:"is_editable"`

	// Request status: OPEN, ACCEPTED, DELIVERED, RECEIVED, COMPLETED, REMOVED
	Status RequestStatus `json:"status"`

	// Profile of the user that created this request
	CreatedBy User `json:"created_by"`

	// Profile of the user that is the provider for this request
	Provider *User `json:"provider"`

	// Users that have offered to carry this request
	PotentialProviders []User `json:"potential_providers"`

	// Organization associated with this request
	Organization Organization `json:"organization"`

	// Visibility restrictions for this request
	Visibility RequestVisibility `json:"visibility"`

	// Short description of item, limited to 255 characters
	Title string `json:"title"`

	// Optional, longer description of the item, limited to 4,096 characters
	Description nulls.String `json:"description"`

	// Geographic location where item is needed
	Destination Location `json:"destination"`

	// Optional geographic location where the item can be picked up, purchased, or otherwise obtained
	Origin *Location `json:"origin"`

	// Broad category of the size of item
	Size string `json:"size"`

	// Date and time this request was created
	CreatedAt time.Time `json:"created_at"`

	// Date and time this request was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Date (yyyy-mm-dd) before which the item will be needed. The record may be hidden or removed after this date
	NeededBefore nulls.Time `json:"needed_before"`

	// Optional weight of the item, measured in kilograms
	Kilograms nulls.Float64 `json:"kilograms"`

	// Optional URL to further describe or point to detail about the item, limited to 255 characters
	URL nulls.String `json:"url"`

	// Photo of the item
	Photo *File `json:"photo"`

	// Meeting associated with this request. Affects visibility of the request.
	Meeting *Meeting `json:"meeting"`
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
	Provider *User `json:"provider"`

	// Short description of item, limited to 255 characters
	Title string `json:"title"`

	// Geographic location where item is needed
	Destination Location `json:"destination"`

	// Optional geographic location where the item can be picked up, purchased, or otherwise obtained
	Origin *Location `json:"origin"`

	// Broad category of the size of item
	Size string `json:"size"`

	// Optional weight of the item, measured in kilograms
	Kilograms nulls.Float64 `json:"kilograms"`

	// Photo of the item
	Photo *File `json:"photo"`
}

// RequestCreateInput includes the fields for creating Requests
//
// swagger:model
type RequestCreateInput struct {
	// Optional, longer description, limited to 4096 characters
	Description nulls.String `json:"description"`

	// Geographic location where item is needed
	Destination Location `json:"destination"`

	// Optional weight of the item, measured in kilograms
	Kilograms nulls.Float64 `json:"kilograms"`

	// Date (yyyy-mm-dd) before which the item will be needed. The record may be hidden or removed after this date.
	NeededBefore nulls.Time `json:"needed_before"`

	// Optional geographic location where the item can be picked up, purchased, or otherwise obtained
	Origin *Location `json:"origin"`

	// ID of associated Organization. Affects visibility of the request, see also the `visibility` field.
	OrganizationID uuid.UUID `json:"org_id"`

	// Optional photo `file` ID. First upload a file using the `/upload` endpoint and then submit its ID here.
	PhotoID nulls.UUID `json:"photo_id"`

	// Broad category of the size of item.
	Size RequestSize `json:"size"`

	// Short description, limited to 255 characters
	Title string `json:"title"`

	// Visibility restrictions for this request, if omitted, the default is "SAME"
	Visibility RequestVisibility `json:"visibility"`
}

// RequestUpdateInput includes the fields for updating Requests
//
// swagger:model
type RequestUpdateInput struct {
	// Optional, longer description, limited to 4096 characters. If omitted or `null`, the description is removed
	Description nulls.String `json:"description"`

	// Geographic location where item is needed. If omitted or `null`, no change is made
	Destination *Location `json:"destination"`

	// Optional weight of the item, measured in kilograms. If omitted or `null`, the value is removed
	Kilograms nulls.Float64 `json:"kilograms"`

	// Date (yyyy-mm-dd) before which the item will be needed. The record may be hidden or removed after this date.
	// If omitted or `null`, the date is removed.
	NeededBefore nulls.Time `json:"needed_before"`

	// Optional geographic location where the item can be picked up, purchased, or otherwise obtained. If omitted or
	// `null`, the origin location is removed.
	Origin *Location `json:"origin"`

	// Optional photo `file` ID. First upload a file using the `/upload` endpoint and then submit its ID here. Any
	// previously attached photo will be deleted. If omitted or `null`, no photo will be attached to this request
	PhotoID nulls.UUID `json:"photo_id"`

	// Broad category of the size of item. If omitted or `null`, no change is made
	Size *RequestSize `json:"size"`

	// Short description, limited to 255 characters.  If omitted or `null`, no change is made.
	Title *string `json:"title"`

	// Visibility restrictions for this request. If omitted or `null`, no change is made.
	Visibility *RequestVisibility `json:"visibility"`
}

// RequestUpdateStatusInput includes the fields for updating the status of a Request
//
// swagger:model
type RequestUpdateStatusInput struct {
	// New Status. Only a limited set of transitions are allowed.
	Status RequestStatus `json:"status"`

	// User ID of the accepted provider. Required if `status` is ACCEPTED and ignored otherwise.
	ProviderUserID *string `json:"provider_user_id"`
}
