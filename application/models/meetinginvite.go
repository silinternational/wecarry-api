package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
)

// MeetingInvite is the model for storing meeting invites sent to prospective users, linked to a meeting/event
type MeetingInvite struct {
	ID        int        `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	MeetingID int        `json:"meeting_id" db:"meeting_id"`
	InviterID int        `json:"inviter_id" db:"inviter_id"`
	Secret    uuid.UUID  `json:"secret" db:"secret"`
	Email     string     `json:"email" db:"email"`
	UserID    nulls.UUID `json:"user_id" db:"user_id"`

	Inviter User    `json:"-" belongs_to:"users" fk_id:"InviterID"`
	Meeting Meeting `json:"-" belongs_to:"meetings" fk_id:"MeetingID"`
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
func (m *MeetingInvite) Create(tx *pop.Connection) error {
	invite := *m
	invite.Secret = domain.GetUUID()

	err := create(tx, &invite)
	if err != nil && strings.Contains(err.Error(), `duplicate key value violates unique constraint`) {
		err = nil
	}
	if err == nil {
		*m = invite

		e := events.Event{
			Kind:    domain.EventApiMeetingInviteCreated,
			Message: "Meeting Invite created",
			Payload: events.Payload{domain.ArgId: m.ID},
		}

		emitEvent(e)
	}

	return err
}

// AvatarURL returns a generated gravatar URL for the inivitee
func (m *MeetingInvite) AvatarURL() string {
	return gravatarURL(m.Email)
}

func (m *MeetingInvite) FindByID(tx *pop.Connection, id int, eagerFields ...string) error {
	if id <= 0 {
		return errors.New("error finding invite: ID must a positive number")
	}

	var err error
	if len(eagerFields) > 0 {
		err = tx.Eager(eagerFields...).Find(m, id)
	} else {
		err = tx.Find(m, id)
	}
	if err != nil {
		return fmt.Errorf("error finding invite by ID: %s", err.Error())
	}

	return nil
}

func (m *MeetingInvite) FindByMeetingIDAndEmail(tx *pop.Connection, meetingID int, email string) error {
	return tx.Where("meeting_id = ? and email = ?", meetingID, email).First(m)
}

func (m *MeetingInvite) Destroy(tx *pop.Connection) error {
	return tx.Destroy(m)
}

// FindBySecret attempts to find a MeetingInvite that exactly matches a given secret, email, and meetingID
func (m *MeetingInvite) FindBySecret(tx *pop.Connection, meetingID int, email, secret string) error {
	if meetingID < 1 {
		return errors.New("invalid meeting ID in FindBySecret")
	}
	if email == "" {
		return errors.New("empty email in FindBySecret")
	}
	if secret == "" {
		return errors.New("empty secret in FindBySecret")
	}
	return tx.Where("meeting_id=? AND email=? AND secret=?", meetingID, email, secret).First(m)
}

func (m *MeetingInvite) InviteURL() string {
	return domain.Env.UIURL + "/invitation?code=" + m.Secret.String()
}

// Updates the MeetingInvite with the UUID of the user who accepted it.
func (m *MeetingInvite) SetUserID(tx *pop.Connection, userID uuid.UUID) error {

	if m.ID == 0 {
		return errors.New("meeting invite must have an id in MeetingInvite.SetUserID")
	}

	if userID == uuid.Nil {
		return errors.New("user uuid must not be zero in MeetingInvite.SetUserID")
	}

	m.UserID = nulls.NewUUID(userID)

	if err := update(tx, m); err != nil {
		return errors.New("error updating meeting invite with a userID: " + err.Error())
	}

	return nil
}
