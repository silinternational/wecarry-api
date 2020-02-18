package models

import (
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
)

// MeetingInvitation is the model for storing meeting invitations sent to prospective users, linked to a meeting/event
type MeetingInvitation struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	MeetingID int       `json:"meeting_id" db:"meeting_id"`
	InviterID int       `json:"inviter_id" db:"inviter_id"`
	Secret    uuid.UUID `json:"secret" db:"secret"`
	Email     string    `json:"email" db:"email"`
}

// MeetingInvitations is used for methods that operate on lists of objects
type MeetingInvitations []MeetingInvitation

// Validate gets run every time you call one of: pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate
func (m *MeetingInvitation) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: m.MeetingID, Name: "MeetingID"},
		&validators.IntIsPresent{Field: m.InviterID, Name: "InviterID"},
		&validators.UUIDIsPresent{Field: m.Secret, Name: "Secret"},
		&validators.EmailIsPresent{Field: m.Email, Name: "Email"},
	), nil
}

// Create validates and stores the MeetingInvitation data as a new record in the database.
func (m *MeetingInvitation) Create() error {
	m.Secret = domain.GetUUID()

	err := create(m)
	if err == nil {
		return err
	}
	if err.Error() == `pq: duplicate key value violates unique constraint "meeting_invitations_meeting_id_email_idx"` {
		return nil
	}
	return err
}

// Meeting returns the related Meeting record
func (m *MeetingInvitation) Meeting() (Meeting, error) {
	var meeting Meeting
	return meeting, DB.Find(&meeting, m.MeetingID)
}

// Inviter returns the related User record of the inviter
func (m *MeetingInvitation) Inviter() (User, error) {
	var user User
	return user, DB.Find(&user, m.InviterID)
}

// AvatarURL returns a generated gravatar URL for the inivitee
func (m *MeetingInvitation) AvatarURL() string {
	return gravatarURL(m.Email)
}
