package api

import (
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/gofrs/uuid"
)

// swagger:model
type Meetings []Meeting

// Meeting a/k/a Event, to serve as a focal point for finding, answering, carrying, and exchanging requests
// swagger:model
type Meeting struct {
	// unique id (uuid) for thread
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Whether meeting is deletable by current user
	// Note: This will only be a valid value when a single
	//       event/meeting is requested (not for a list of events)
	IsDeletable nulls.Bool `json:"is_deletable"`

	// Whether meeting is editable by current user
	IsEditable bool `json:"is_editable"`

	// Short name, limited to 80 characters
	Name string `json:"name"`

	// the text-only description, limited to 4096 characters"
	Description string `json:"description"`

	// The date of the first day of the meeting (event)
	StartDate string `json:"start_date"`

	// The date of the last day of the meeting (event)
	EndDate string `json:"end_date"`

	// Profile of the user that added this meeting (event) to the app
	CreatedBy User `json:"created_by"`

	// Image File -- typically a logo
	ImageFile *File `json:"image_file,omitempty"`

	// invites for this meeting
	Invites MeetingInvites `json:"invites"`

	// location -- notifications and filters may use this location
	Location Location `json:"location,omitempty"`

	// Whether the current user has joined this event/meeting
	HasJoined bool `json:"has_joined"`

	// participants
	Participants MeetingParticipants `json:"participants"`

	// The time the meeting (event) was added to the app"
	CreatedAt time.Time `json:"created_at"`

	// The time the meeting (event) was last modified in the app"
	UpdatedAt time.Time `json:"updated_at"`

	// meeting (event) information URL -- should be a full website, but could be an information document such as a pdf"
	MoreInfoURL string `json:"more_info_url"`
}

// MeetingInput includes the fields for creating or updating Meetings/Events
//
// swagger:model
type MeetingInput struct {

	// short name, limited to 80 characters
	Name string `json:"name"`

	// text-only description, limited to 4096 characters
	Description nulls.String `json:"description"`

	// date (yyyy-mm-dd) of the first day of the meeting (event)"
	StartDate string `json:"start_date"`

	// date (yyyy-mm-dd) of the last day of the meeting (event)"
	EndDate string `json:"end_date"`

	// meeting (event) information URL -- should be a full website, but could be an information document such as a pdf"
	MoreInfoURL nulls.String `json:"more_info_url"`

	// ID of pre-stored image file, typically a logo. Upload using the `upload` REST API endpoint."
	ImageFileID nulls.UUID `json:"image_file_id"`

	// meeting (event) location -- notifications and filters may use this location"
	Location Location `json:"location"`

	// email addresses to which to send meeting invites. Can be comma- or newline-separated.
	Emails string `json:"emails"`
}

// swagger:model
type MeetingParticipants []MeetingParticipant

// Confirmed participant of a `Meeting`. An invited person will not appear as a `MeetingParticipant`
//   until they have confirmed a `MeetingInvite` or self-joined a non-INVITE_ONLY meeting.
//
// swagger:model
type MeetingParticipant struct {
	// `User` information for the `Meeting` participant
	User User `json:"user"`

	// "true if `User` is a meeting Organizer"
	IsOrganizer bool `json:"is_organizer"`
}

// MeetingParticipantInput contains parameters to join an event/meeting by creating a MessageParticipant
// swagger:model
type MeetingParticipantInput struct {
	// ID of the `Meeting`
	MeetingID string `json:"meeting_id"`

	// Secret code from the `MeetingInvite` or invite code from the `Meeting`.
	// If the `Meeting` is not `INVITE_ONLY`, the code may be omitted.
	Code *string `json:"code"`
}

// swagger:model
type MeetingInvites []MeetingInvite

// Invitation to a `Meeting`.
//
// swagger:model
type MeetingInvite struct {
	// The uuid of the meeting the invite is for
	MeetingID uuid.UUID `json:"meeting_id"`

	// The uuid of the user who created the invite
	InviterID uuid.UUID `json:"inviter_id"`

	// The email address of the person being invited
	Email string `json:"email" db:"email"`
}

// The email address associated with a meeting invite
//
// swagger:model
type MeetingInviteEmail struct {
	InviteEmail string `json:"invite_email"`
}
