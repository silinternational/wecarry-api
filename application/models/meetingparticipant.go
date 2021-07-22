package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/silinternational/wecarry-api/api"

	"github.com/silinternational/wecarry-api/domain"
)

// MeetingParticipant is the model for storing meeting participants, users linked to a meeting/event
type MeetingParticipant struct {
	ID          int       `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	MeetingID   int       `json:"meeting_id" db:"meeting_id"`
	UserID      int       `json:"user_id" db:"user_id"`
	InviteID    nulls.Int `json:"invite_id" db:"invite_id"`
	IsOrganizer bool      `json:"is_organizer" db:"is_organizer"`
}

// MeetingParticipants is used for methods that operate on lists of objects
type MeetingParticipants []MeetingParticipant

// String is used to serialize the object for error logging
func (m MeetingParticipant) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call one of: pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate
func (m *MeetingParticipant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: m.MeetingID, Name: "MeetingID"},
		&validators.IntIsPresent{Field: m.UserID, Name: "UserID"},
	), nil
}

// Meeting returns the related Meeting record
func (m *MeetingParticipant) Meeting(tx *pop.Connection) (Meeting, error) {
	var meeting Meeting
	return meeting, tx.Find(&meeting, m.MeetingID)
}

// User returns the related User record of the participant
func (m *MeetingParticipant) User(tx *pop.Connection) (User, error) {
	var user User
	return user, tx.Find(&user, m.UserID)
}

// Invite returns the related MeetingInvite record
func (m *MeetingParticipant) Invite(tx *pop.Connection) (*MeetingInvite, error) {
	var invite MeetingInvite
	if !m.InviteID.Valid {
		return nil, nil
	}
	err := tx.Find(&invite, m.InviteID.Int)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

// FindByMeetingIDAndUserID does what it says
func (m *MeetingParticipant) FindByMeetingIDAndUserID(tx *pop.Connection, meetingID, userID int) error {
	return tx.Where("meeting_id = ? and user_id = ?", meetingID, userID).First(m)
}

// CreateFromInvite creates a new MeetingParticipant using the MeetingInvite's information
func (m *MeetingParticipant) CreateFromInvite(tx *pop.Connection, invite MeetingInvite, userID int) error {
	m.InviteID = nulls.NewInt(invite.ID)
	m.UserID = userID
	m.MeetingID = invite.MeetingID
	return tx.Create(m)
}

func (m *MeetingParticipant) Destroy(tx *pop.Connection) error {
	return tx.Destroy(m)
}

// FindOrCreate a new MeetingParticipant from a meeting ID and code. If `code` is nil, the meeting must be non-INVITE_ONLY.
// Otherwise, `code` must match either a MeetingInvite secret code or a Meeting invite code.
func (m *MeetingParticipant) FindOrCreate(tx *pop.Connection, meeting Meeting, user User, code *string) *api.AppError {
	if err := m.FindByMeetingIDAndUserID(tx, meeting.ID, user.ID); domain.IsOtherThanNoRows(err) {
		return &api.AppError{
			Key: "CreateMeetingParticipant.FindExisting",
			Err: err,
		}
	}
	if m.ID > 0 {
		return nil
	}

	if code == nil {
		if user.CanCreateMeetingParticipant(tx, meeting) {
			return m.createWithoutInvite(tx, user, meeting)
		}
		return &api.AppError{
			Err: errors.New("user is not allowed to self-join meeting without a code"),
			Key: "CreateMeetingParticipant.Unauthorized",
		}
	}

	if meeting.IsCodeValid(tx, *code) {
		return m.createWithoutInvite(tx, user, meeting)
	}

	var invite MeetingInvite
	if err := invite.FindBySecret(tx, meeting.ID, user.Email, *code); domain.IsOtherThanNoRows(err) {
		return &api.AppError{
			Err: err,
			Key: "CreateMeetingParticipant.FindBySecret",
		}
	}
	if invite.ID == 0 {
		return &api.AppError{
			Err: errors.New("invalid invite secret"),
			Key: "CreateMeetingParticipant.InvalidSecret",
		}
	}

	m.InviteID = nulls.NewInt(invite.ID)
	m.UserID = user.ID
	m.MeetingID = meeting.ID
	if err := tx.Create(m); err != nil {
		return &api.AppError{
			Err: err,
			Key: "CreateMeetingParticipant",
		}
	}
	return nil
}

func (m *MeetingParticipant) createWithoutInvite(tx *pop.Connection, user User, meeting Meeting) *api.AppError {
	m.UserID = user.ID
	m.MeetingID = meeting.ID
	if err := tx.Create(m); err != nil {
		return &api.AppError{
			Key: "CreateMeetingParticipant",
			Err: err,
		}
	}
	return nil
}
