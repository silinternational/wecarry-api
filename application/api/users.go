package api

import (
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

// UserPrivate has the full set of User attributes NOT intended for public exposure
// swagger:model
type UserPrivate struct {
	// unique identifier for the User
	// swagger:strfmt uuid4
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Email address to be used for notifications to the User. Not necessarily the same as the authentication email.
	Email string `json:"email"`

	// User's nickname. Auto-assigned upon creation of a User, but editable by the User. Limited to 255 characters.
	Nickname string `json:"nickname"`

	// `File` ID of the user's photo, if present
	// swagger:strfmt uuid4
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	PhotoID uuid.UUID `json:"photo_id"`

	// avatarURL is generated from an attached photo if present, an external URL if present, or a Gravatar URL
	// swagger:strfmt url
	AvatarURL nulls.String `json:"avatar_url"`

	// Organizations that the User is affilated with. This can be empty or have a single entry. Future capability is TBD
	Organizations []Organization `json:"organizations"`
}

// swagger:model
type Users []User

// User has a limited set of User attributes which are intended for public exposure
// swagger:model
type User struct {
	// unique identifier for the User
	// swagger:strfmt uuid4
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// User's nickname. Auto-assigned upon creation of a User, but editable by the User. Limited to 255 characters.
	Nickname string `json:"nickname"`

	// avatarURL is generated from an attached photo if present, an external URL if present, or a Gravatar URL
	// swagger:strfmt url
	AvatarURL nulls.String `json:"avatar_url"`
}
