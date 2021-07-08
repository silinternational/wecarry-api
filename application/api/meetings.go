package api

import (
	"time"

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

	// Short name, limited to 80 characters
	Name string `json:"name"`

	// the text-only description, limited to 4096 characters"
	Description string `json:"description"`

	// The date of the first day of the meeting (event)
	StartDate time.Time `json:"start_date"`

	// The date of the last day of the meeting (event)
	EndDate time.Time `json:"end_date"`

	// Profile of the user that added this meeting (event) to the app
	CreatedBy User `json:"created_by"`

	// Image File -- typically a logo
	ImageFile *File `json:"image_file,omitempty"`

	// location -- notifications and filters may use this location
	Location Location `json:"location,omitempty"`

	// participants
	Participants MeetingParticipants `json:"participants"`

	// The time the meeting (event) was added to the app"
	CreatedAt time.Time `json:"created_at"`

	// The time the meeting (event) was last modified in the app"
	UpdatedAt time.Time `json:"updated_at"`

	// meeting (event) information URL -- should be a full website, but could be an information document such as a pdf"
	MoreInfoURL string `json:"more_info_url"`
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
