package models

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"

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

// FindByMeetingIDAndUserID does what it says
func (m *MeetingParticipant) FindByMeetingIDAndUserID(meetingID, userID int) error {
	return DB.Where("meeting_id = ? and user_id = ?", meetingID, userID).First(m)
}

// CreateFromInvite creates a new MeetingParticipant using the MeetingInvite's information
func (m *MeetingParticipant) CreateFromInvite(invite MeetingInvite, userID int) error {
	m.InviteID = nulls.NewInt(invite.ID)
	m.UserID = userID
	m.MeetingID = invite.MeetingID
	return DB.Create(m)
}

func (m *MeetingParticipant) Destroy() error {
	return DB.Destroy(m)
}

// Create a new MeetingParticipant from a meeting ID and code. If `code` is nil, the meeting must be non-INVITE_ONLY.
// Otherwise, `code` must match either a MeetingInvite secret code or a Meeting invite code.
func (m *MeetingParticipant) Create(ctx context.Context, meeting Meeting, code *string) error {
	cUser := CurrentUser(ctx)

	var p MeetingParticipant
	if err := p.FindByMeetingIDAndUserID(meeting.ID, cUser.ID); domain.IsOtherThanNoRows(err) {
		return domain.ReportError(ctx, err,
			"CreateMeetingParticipant.FindExisting")
	}
	if p.ID > 0 {
		return nil
	}

	if code == nil {
		if cUser.CanCreateMeetingParticipant(domain.GetBuffaloContext(ctx), meeting) {
			return m.createWithoutInvite(ctx, meeting)
		}
		return domain.ReportError(ctx, errors.New("user is not allowed to self-join meeting without a code"),
			"CreateMeetingParticipant.Unauthorized")
	}

	if meeting.IsCodeValid(*code) {
		return m.createWithoutInvite(ctx, meeting)
	}

	var invite MeetingInvite
	if err := invite.FindBySecret(meeting.ID, cUser.Email, *code); domain.IsOtherThanNoRows(err) {
		return domain.ReportError(ctx, err, "CreateMeetingParticipant.FindBySecret")
	}
	if invite.ID == 0 {
		return domain.ReportError(ctx, errors.New("invalid invite secret"), "CreateMeetingParticipant.InvalidSecret")
	}

	m.InviteID = nulls.NewInt(invite.ID)
	m.UserID = cUser.ID
	m.MeetingID = meeting.ID
	if err := DB.Create(m); err != nil {
		return domain.ReportError(ctx, err, "CreateMeetingParticipant")
	}
	return nil
}

func (m *MeetingParticipant) createWithoutInvite(ctx context.Context, meeting Meeting) error {
	m.UserID = CurrentUser(ctx).ID
	m.MeetingID = meeting.ID
	if err := DB.Create(m); err != nil {
		return domain.ReportError(ctx, err, "CreateMeetingParticipant")
	}
	return nil
}
