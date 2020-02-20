package models

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
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

// Validate gets run every time you call one of: pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate
func (m *MeetingParticipant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: m.MeetingID, Name: "MeetingID"},
		&validators.IntIsPresent{Field: m.UserID, Name: "UserID"},
	), nil
}

// Meeting returns the related Meeting record
func (m *MeetingParticipant) Meeting() (Meeting, error) {
	var meeting Meeting
	return meeting, DB.Find(&meeting, m.MeetingID)
}

// User returns the related User record of the participant
func (m *MeetingParticipant) User() (User, error) {
	var user User
	return user, DB.Find(&user, m.UserID)
}

// Invite returns the related MeetingInvite record
func (m *MeetingParticipant) Invite() (*MeetingInvite, error) {
	var invite MeetingInvite
	if !m.InviteID.Valid {
		return nil, nil
	}
	err := DB.Find(&invite, m.InviteID.Int)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (m *MeetingParticipant) FindByMeetingIDAndUser(meetingID, userID int) error {
	return DB.Where("meeting_id = ? and user_id = ?", meetingID, userID).First(m)
}

func (m *MeetingParticipant) CreateForInvite(invite MeetingInvite, userID int) error {
	m.InviteID = nulls.NewInt(invite.ID)
	m.UserID = userID
	m.MeetingID = invite.MeetingID
	return DB.Create(m)
}
