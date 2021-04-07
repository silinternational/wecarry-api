package models

import (
	"errors"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
)

// MeetingInvite is the model for storing meeting invites sent to prospective users, linked to a meeting/event
type MeetingInvite struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	MeetingID int       `json:"meeting_id" db:"meeting_id"`
	InviterID int       `json:"inviter_id" db:"inviter_id"`
	Secret    uuid.UUID `json:"secret" db:"secret"`
	Email     string    `json:"email" db:"email"`
}

// MeetingInvites is used for methods that operate on lists of objects
type MeetingInvites []MeetingInvite

// Validate gets run every time you call one of: pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate
func (m *MeetingInvite) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: m.MeetingID, Name: "MeetingID"},
		&validators.IntIsPresent{Field: m.InviterID, Name: "InviterID"},
		&validators.UUIDIsPresent{Field: m.Secret, Name: "Secret"},
		&validators.EmailIsPresent{Field: m.Email, Name: "Email"},
	), nil
}

// Create validates and stores the MeetingInvite data as a new record in the database.
func (m *MeetingInvite) Create() error {
	invite := *m
	invite.Secret = domain.GetUUID()

	err := create(&invite)
	if err != nil && strings.Contains(err.Error(), `duplicate key value violates unique constraint`) {
		err = nil
	}
	if err == nil {
		*m = invite
	}
	return err
}

// Meeting returns the related Meeting record
func (m *MeetingInvite) Meeting() (Meeting, error) {
	var meeting Meeting
	return meeting, DB.Find(&meeting, m.MeetingID)
}

// Inviter returns the related User record of the inviter
func (m *MeetingInvite) Inviter() (User, error) {
	var user User
	return user, DB.Find(&user, m.InviterID)
}

// AvatarURL returns a generated gravatar URL for the inivitee
func (m *MeetingInvite) AvatarURL() string {
	return gravatarURL(m.Email)
}

func (m *MeetingInvite) FindByMeetingIDAndEmail(meetingID int, email string) error {
	return DB.Where("meeting_id = ? and email = ?", meetingID, email).First(m)
}

func (m *MeetingInvite) Destroy() error {
	return DB.Destroy(m)
}

// FindBySecret attempts to find a MeetingInvite that exactly matches a given secret, email, and meetingID
func (m *MeetingInvite) FindBySecret(meetingID int, email, secret string) error {
	if meetingID < 1 {
		return errors.New("invalid meeting ID in FindBySecret")
	}
	if email == "" {
		return errors.New("empty email in FindBySecret")
	}
	if secret == "" {
		return errors.New("empty secret in FindBySecret")
	}
	return DB.Where("meeting_id=? AND email=? AND secret=?", meetingID, email, secret).First(m)
}
