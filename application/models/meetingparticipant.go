package models

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"

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
func (m *MeetingParticipant) Meeting(ctx context.Context) (Meeting, error) {
	var meeting Meeting
	return meeting, Tx(ctx).Find(&meeting, m.MeetingID)
}

// User returns the related User record of the participant
func (m *MeetingParticipant) User(ctx context.Context) (User, error) {
	var user User
	return user, Tx(ctx).Find(&user, m.UserID)
}

// Invite returns the related MeetingInvite record
func (m *MeetingParticipant) Invite(ctx context.Context) (*MeetingInvite, error) {
	var invite MeetingInvite
	if !m.InviteID.Valid {
		return nil, nil
	}
	err := Tx(ctx).Find(&invite, m.InviteID.Int)
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
func (m *MeetingParticipant) FindOrCreate(ctx context.Context, meeting Meeting, code *string) error {
	cUser := CurrentUser(ctx)
	tx := Tx(ctx)
	if err := m.FindByMeetingIDAndUserID(tx, meeting.ID, cUser.ID); domain.IsOtherThanNoRows(err) {
		return domain.ReportError(ctx, err,
			"CreateMeetingParticipant.FindExisting")
	}
	if m.ID > 0 {
		return nil
	}

	if code == nil {
		if cUser.CanCreateMeetingParticipant(ctx, meeting) {
			return m.createWithoutInvite(ctx, meeting)
		}
		return domain.ReportError(ctx, errors.New("user is not allowed to self-join meeting without a code"),
			"CreateMeetingParticipant.Unauthorized")
	}

	if meeting.IsCodeValid(tx, *code) {
		return m.createWithoutInvite(ctx, meeting)
	}

	var invite MeetingInvite
	if err := invite.FindBySecret(tx, meeting.ID, cUser.Email, *code); domain.IsOtherThanNoRows(err) {
		return domain.ReportError(ctx, err, "CreateMeetingParticipant.FindBySecret")
	}
	if invite.ID == 0 {
		return domain.ReportError(ctx, errors.New("invalid invite secret"), "CreateMeetingParticipant.InvalidSecret")
	}

	m.InviteID = nulls.NewInt(invite.ID)
	m.UserID = cUser.ID
	m.MeetingID = meeting.ID
	if err := tx.Create(m); err != nil {
		return domain.ReportError(ctx, err, "CreateMeetingParticipant")
	}
	return nil
}

func (m *MeetingParticipant) createWithoutInvite(ctx context.Context, meeting Meeting) error {
	m.UserID = CurrentUser(ctx).ID
	m.MeetingID = meeting.ID
	if err := Tx(ctx).Create(m); err != nil {
		return domain.ReportError(ctx, err, "CreateMeetingParticipant")
	}
	return nil
}
