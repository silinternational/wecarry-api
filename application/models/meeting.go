package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
)

const (
	OptIncludeParticipants MeetingOption = iota + 1
	OptIncludeInvites      MeetingOption = iota
)

type MeetingOption int

func (o *MeetingOption) isSelected(options ...MeetingOption) bool {
	if len(options) == 0 {
		return false
	}

	for _, mo := range options {
		if *o == mo {
			return true
		}
	}

	return false
}

// Meeting represents an event where people gather together from different locations
type Meeting struct {
	ID          int          `json:"-" db:"id"`
	UUID        uuid.UUID    `json:"uuid" db:"uuid"`
	Name        string       `json:"name" db:"name"`
	Description nulls.String `json:"description" db:"description"`
	MoreInfoURL nulls.String `json:"more_info_url" db:"more_info_url"`
	StartDate   time.Time    `json:"start_date" db:"start_date"`
	EndDate     time.Time    `json:"end_date" db:"end_date"`
	InviteCode  nulls.UUID   `json:"invite_code" db:"invite_code"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	CreatedByID int          `json:"created_by_id" db:"created_by_id"`
	FileID      nulls.Int    `json:"file_id" db:"file_id"`
	LocationID  int          `json:"location_id" db:"location_id"`

	CreatedBy User     `json:"-" belongs_to:"users" fk_id:"CreatedByID"`
	ImgFile   *File    `json:"-" belongs_to:"files" fk_id:"FileID"`
	Location  Location `json:"-" belongs_to:"locations"`
}

// String is not required by pop and may be deleted
func (m Meeting) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Meetings is not required by pop and may be deleted
type Meetings []Meeting

// String is not required by pop and may be deleted
func (m Meetings) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *Meeting) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: m.UUID, Name: "UUID"},
		&validators.StringIsPresent{Field: m.Name, Name: "Name"},
		&validators.TimeIsPresent{Field: m.StartDate, Name: "StartDate"},
		&validators.TimeIsPresent{Field: m.EndDate, Name: "EndDate"},
		&validators.IntIsPresent{Field: m.CreatedByID, Name: "CreatedByID"},
		&validators.IntIsPresent{Field: m.LocationID, Name: "LocationID"},
		&dateValidator{StartDate: m.StartDate, EndDate: m.EndDate, Name: "Dates"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (m *Meeting) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (m *Meeting) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

type dateValidator struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Message   string
}

func (v *dateValidator) IsValid(errors *validate.Errors) {
	if v.StartDate.Before(v.EndDate) {
		return
	}

	if v.StartDate.Format(domain.DateFormat) == v.EndDate.Format(domain.DateFormat) {
		return
	}

	v.Message = fmt.Sprintf("Start date must come no later than end date chronologically. Got %s and %s.",
		v.StartDate.Format(domain.DateFormat), v.EndDate.Format(domain.DateFormat))
	errors.Add(validators.GenerateKey(v.Name), v.Message)
}

func (m *Meeting) SafeDelete(tx *pop.Connection) error {
	requests, err := m.Requests(tx)
	if domain.IsOtherThanNoRows(err) {
		return err
	}

	if len(requests) > 0 {
		return errors.New("meeting with associated requests may not be deleted")
	}

	return tx.Destroy(m)
}

// FindByUUID finds a meeting by the UUID field and loads its CreatedBy field
func (m *Meeting) FindByUUID(tx *pop.Connection, uuid string) error {
	if uuid == "" {
		return errors.New("error finding meeting: uuid must not be blank")
	}

	if err := tx.Where("uuid = ?", uuid).First(m); err != nil {
		return fmt.Errorf("error finding meeting by uuid: %s", err.Error())
	}

	return nil
}

func getOrdered(m *Meetings, q *pop.Query) error {
	return q.Order("start_date asc").All(m)
}

// FindOnDate finds the meetings that have StartDate before timeInFocus-date and an EndDate after it
// (inclusive on both)
func (m *Meetings) FindOnDate(tx *pop.Connection, timeInFocus time.Time) error {
	date := timeInFocus.Format(domain.DateTimeFormat)
	where := "start_date <= ? and end_date >= ?"

	if err := getOrdered(m, tx.Where(where, date, date)); err != nil {
		return fmt.Errorf("error finding meeting with start_date and end_date straddling %s ... %s",
			date, err.Error())
	}

	return nil
}

// FindOnOrAfterDate finds the meetings that have an EndDate on or after the timeInFocus-date
func (m *Meetings) FindOnOrAfterDate(tx *pop.Connection, timeInFocus time.Time) error {
	date := timeInFocus.Format(domain.DateTimeFormat)

	if err := getOrdered(m, tx.Where("end_date >= ?", date)); err != nil {
		return fmt.Errorf("error finding meeting with end_date before %s ... %s", date, err.Error())
	}

	return nil
}

// FindAfterDate finds the meetings that have a StartDate after the timeInFocus-date
func (m *Meetings) FindAfterDate(tx *pop.Connection, timeInFocus time.Time) error {
	date := timeInFocus.Format(domain.DateTimeFormat)

	if err := getOrdered(m, tx.Where("start_date > ?", date)); err != nil {
		return fmt.Errorf("error finding meeting with start_date after %s ... %s", date, err.Error())
	}

	return nil
}

// FindRecent finds the meetings that have an EndDate within the past <domain.RecentMeetingDelay> days
// before timeInFocus-date (not inclusive)
func (m *Meetings) FindRecent(tx *pop.Connection, timeInFocus time.Time) error {
	yesterday := timeInFocus.Add(-domain.DurationDay).Format(domain.DateTimeFormat)
	recentDate := timeInFocus.Add(-domain.RecentMeetingDelay)
	where := "end_date between ? and ?"

	if err := getOrdered(m, tx.Where(where, recentDate, yesterday)); err != nil {
		return fmt.Errorf("error finding meeting with end_date between %s and %s ... %s",
			recentDate, yesterday, err.Error())
	}

	return nil
}

func (m *Meeting) FindByInviteCode(tx *pop.Connection, code string) error {
	if code == "" {
		return errors.New("error finding meeting: invite_code must not be blank")
	}

	if err := tx.Where("invite_code = ?", code).First(m); err != nil {
		return fmt.Errorf("error finding meeting by invite_code: %s", err.Error())
	}

	return nil
}

// SetImageFile assigns a previously-stored File to this Meeting as its image. Parameter `fileID` is the UUID
// of the image to attach.
func (m *Meeting) SetImageFile(tx *pop.Connection, fileID string) (File, error) {
	return addFile(tx, m, fileID)
}

// ImageFile retrieves the file attached as the Meeting Image
func (m *Meeting) ImageFile(tx *pop.Connection) (*File, error) {
	if !m.FileID.Valid {
		return nil, nil
	}
	if m.ImgFile == nil {
		if err := tx.Load(m, "ImgFile"); err != nil {
			return nil, err
		}
	}
	if err := (*m.ImgFile).RefreshURL(tx); err != nil {
		return nil, err
	}
	f := *m.ImgFile
	return &f, nil
}

// RemoveFile removes an attached file from the Meeting
func (m *Meeting) RemoveFile(tx *pop.Connection) error {
	return removeFile(tx, m)
}

func (m *Meeting) GetCreator(tx *pop.Connection) (*User, error) {
	creator := User{}
	if err := tx.Find(&creator, m.CreatedByID); err != nil {
		return nil, err
	}
	return &creator, nil
}

// GetLocation returns the related Location object.
func (m *Meeting) GetLocation(tx *pop.Connection) (Location, error) {
	location := Location{}
	if err := tx.Find(&location, m.LocationID); err != nil {
		return location, err
	}

	return location, nil
}

// Create stores the Meeting data as a new record in the database.
func (m *Meeting) Create(tx *pop.Connection) error {
	return create(tx, m)
}

// Update writes the Meeting data to an existing database record.
func (m *Meeting) Update(tx *pop.Connection) error {
	return update(tx, m)
}

// SetLocation sets the location field, creating a new record in the database if necessary.
func (m *Meeting) SetLocation(tx *pop.Connection, location Location) error {
	location.ID = m.LocationID
	if m.LocationID == 0 {
		if err := location.Create(tx); err != nil {
			return err
		}
	} else {
		if err := location.Update(tx); err != nil {
			return err
		}
	}
	m.LocationID = location.ID
	m.Location = location

	return nil
}

// CanCreate returns a bool based on whether the current user is allowed to create a meeting
func (m *Meeting) CanCreate(user User) bool {
	return true
}

// CanDelete returns a bool based on whether the current user is
//   allowed to update this meeting and the meeting has no requests
//   associated with it.
func (m *Meeting) CanDelete(tx *pop.Connection, user User) (bool, error) {
	if !user.CanUpdateMeeting(*m) {
		return false, nil
	}

	requests, err := m.Requests(tx)
	if domain.IsOtherThanNoRows(err) {
		return false, err
	}

	return len(requests) == 0, nil
}

// CanUpdate returns a bool based on whether the current user is allowed to update a meeting
func (m *Meeting) CanUpdate(user User) bool {
	switch user.AdminRole {
	case UserAdminRoleSuperAdmin, UserAdminRoleSalesAdmin, UserAdminRoleAdmin:
		return true
	}

	return user.ID == m.CreatedByID
}

// Requests return all associated Requests
func (m *Meeting) Requests(tx *pop.Connection) (Requests, error) {
	var requests Requests
	if err := tx.Where("meeting_id = ?", m.ID).Order("updated_at desc").All(&requests); err != nil {
		return nil, fmt.Errorf("error getting requests for meeting id %v ... %v", m.ID, err)
	}

	return requests, nil
}

// Invites returns all of the MeetingInvites for this Meeting. Only the meeting creator and organizers are authorized.
func (m *Meeting) Invites(tx *pop.Connection, user User) (MeetingInvites, error) {
	i := MeetingInvites{}
	if m == nil {
		return i, nil
	}

	isOrganizer, err := m.isOrganizer(tx, user.ID)
	if err != nil {
		return i, err
	}

	if user.ID != m.CreatedByID && !isOrganizer && !user.isSuperAdmin() {
		return i, nil
	}
	if err := tx.Where("meeting_id = ?", m.ID).Eager("Inviter").All(&i); err != nil {
		return i, err
	}
	return i, nil
}

// Participants returns all of the MeetingParticipants for this Meeting. Only the meeting creator and organizers are
// authorized.
func (m *Meeting) Participants(tx *pop.Connection, user User) (MeetingParticipants, error) {
	p := MeetingParticipants{}
	if m == nil {
		return p, nil
	}

	isOrganizer, err := m.isOrganizer(tx, user.ID)
	if err != nil {
		return p, err
	}

	if user.ID != m.CreatedByID && !isOrganizer && !user.isSuperAdmin() {
		return p, tx.Where("user_id = ? AND meeting_id = ?", user.ID, m.ID).All(&p)
	}
	if err := tx.Where("meeting_id = ?", m.ID).All(&p); err != nil {
		return p, err
	}
	return p, nil
}

// Organizers returns all of the users who are organizers for this Meeting. No authorization is checked, so
// any queries should render this as a PublicProfile to limit field visibility.
func (m *Meeting) Organizers(tx *pop.Connection) (Users, error) {
	u := Users{}
	if m == nil {
		return u, nil
	}
	if err := tx.
		Select("users.id", "users.uuid", "nickname", "file_id", "auth_photo_url").
		Where("meeting_participants.is_organizer=true").
		Where("meeting_participants.meeting_id=?", m.ID).
		Join("meeting_participants", "meeting_participants.user_id=users.id").
		All(&u); err != nil {
		return u, err
	}
	return u, nil
}

func (m *Meeting) RemoveInvite(tx *pop.Connection, email string) error {
	var invite MeetingInvite
	if err := invite.FindByMeetingIDAndEmail(tx, m.ID, email); err != nil {
		return err
	}
	return invite.Destroy(tx)
}

func (m *Meeting) RemoveParticipant(tx *pop.Connection, userUUID string) error {
	var user User
	if err := user.FindByUUID(tx, userUUID); err != nil {
		return fmt.Errorf("invalid user ID %s in Meeting.RemoveParticipant, %s", userUUID, err)
	}
	var participant MeetingParticipant
	if err := participant.FindByMeetingIDAndUserID(tx, m.ID, user.ID); err != nil {
		return fmt.Errorf("failed to load MeetingParticipant in Meeting.RemoveParticipant, %s", err)
	}
	return participant.Destroy(tx)
}

func (m *Meeting) IsCodeValid(tx *pop.Connection, code string) bool {
	if m.InviteCode.Valid && m.InviteCode.UUID.String() == code {
		return true
	}
	return false
}

func (m *Meeting) isOrganizer(tx *pop.Connection, userID int) (bool, error) {
	organizers, err := m.Organizers(tx)
	if err != nil {
		return false, errors.New("isOrganizer() error reading list of meeting organizers, " + err.Error())
	}
	for _, o := range organizers {
		if o.ID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *Meeting) isVisible(tx *pop.Connection, userID int) bool {
	return true
}

// FindByIDs finds all Meetings associated with the given IDs and loads them from the database
func (m *Meetings) FindByIDs(tx *pop.Connection, ids []int) error {
	ids = domain.UniquifyIntSlice(ids)
	return tx.Where("id in (?)", ids).All(m)
}

// ConvertMeetings converts list of models.Meeting into list of api.Meeting
func ConvertMeetings(ctx context.Context, meetings []Meeting, user User) ([]api.Meeting, error) {
	output := make([]api.Meeting, len(meetings))

	for i, m := range meetings {
		var err error
		output[i], err = ConvertMeeting(ctx, m, user)
		if err != nil {
			return []api.Meeting{}, err
		}
	}

	return output, nil
}

// ConvertMeeting converts a model.Meeting into api.Meeting
func ConvertMeeting(ctx context.Context, meeting Meeting, user User, options ...MeetingOption) (api.Meeting, error) {
	output := convertMeetingAbridged(meeting)
	tx := Tx(ctx)
	if err := tx.Load(&meeting); err != nil {
		return api.Meeting{}, err
	}

	createdBy, err := loadMeetingCreatedBy(ctx, meeting)
	if err != nil {
		return api.Meeting{}, err
	}
	output.CreatedBy = createdBy

	output.ImageFile = convertMeetingImageFile(meeting)
	output.Location = convertLocation(meeting.Location)
	output.HasJoined = true

	var userP MeetingParticipant
	if err := userP.FindByMeetingIDAndUserID(tx, meeting.ID, user.ID); err != nil {
		if domain.IsOtherThanNoRows(err) {
			err := fmt.Errorf("failed to load MeetingParticipant in Meeting.ConvertMeeting, %s", err)
			return api.Meeting{}, err
		}
		output.HasJoined = false // no participant found for user
	}

	if err := loadMeetingOptions(ctx, meeting, user, &output, options...); err != nil {
		return api.Meeting{}, err
	}

	output.IsEditable = meeting.CanUpdate(user)

	return output, nil
}

func loadMeetingOptions(ctx context.Context, meeting Meeting, user User, apiMeeting *api.Meeting, options ...MeetingOption) error {
	apiMeeting.Participants = api.MeetingParticipants{}

	opt := OptIncludeParticipants
	if opt.isSelected(options...) {
		participants, err := loadMeetingParticipants(ctx, meeting, user)
		if err != nil {
			return err
		}
		apiMeeting.Participants = participants
	}

	apiMeeting.Invites = api.MeetingInvites{}
	opt = OptIncludeInvites
	if opt.isSelected(options...) {
		invites, err := loadMeetingInvites(Tx(ctx), meeting, user)
		if err != nil {
			return err
		}
		apiMeeting.Invites = invites
	}

	return nil
}

// convertMeetingImageFile converts a model.Meeting.ImgFile into an api.File
// This assumes that the meeting's related ImageFile has already been loaded
func convertMeetingImageFile(meeting Meeting) *api.File {
	if meeting.ImgFile == nil {
		return nil
	}
	file := convertFile(*meeting.ImgFile)
	return &file
}

// convertMeetingParticipants converts model.MeetingParticipants into an api.MeetingParticipants
func convertMeetingParticipants(ctx context.Context, participants MeetingParticipants) (api.MeetingParticipants, error) {
	output := make(api.MeetingParticipants, len(participants))
	for i := range output {
		var err error
		output[i], err = convertMeetingParticipant(ctx, participants[i])
		if err != nil {
			return output, err
		}
	}
	return output, nil
}

func convertMeetingParticipant(ctx context.Context, participant MeetingParticipant) (api.MeetingParticipant, error) {
	tx := Tx(ctx)

	output := api.MeetingParticipant{}

	user, err := participant.User(tx)
	if err != nil {
		return api.MeetingParticipant{}, err
	}

	outputUser, err := ConvertUser(ctx, user)
	if err != nil {
		return api.MeetingParticipant{}, err
	}
	output.User = outputUser

	output.IsOrganizer = participant.IsOrganizer

	return output, nil
}

func convertMeetingAbridged(meeting Meeting) api.Meeting {
	return api.Meeting{
		ID:          meeting.UUID,
		Name:        meeting.Name,
		Description: meeting.Description.String,
		StartDate:   meeting.StartDate.Format(domain.DateFormat),
		EndDate:     meeting.EndDate.Format(domain.DateFormat),
		CreatedAt:   meeting.CreatedAt,
		UpdatedAt:   meeting.UpdatedAt,
		MoreInfoURL: meeting.MoreInfoURL.String,
	}
}

func loadMeetingCreatedBy(ctx context.Context, meeting Meeting) (api.User, error) {
	outputCreatedBy, err := ConvertUser(ctx, meeting.CreatedBy)
	if err != nil {
		err = errors.New("error converting meeting created_by user: " + err.Error())
		return api.User{}, err
	}
	return outputCreatedBy, nil
}

func loadMeetingImageFile(ctx context.Context, meeting Meeting) (*api.File, error) {
	imageFile, err := meeting.ImageFile(Tx(ctx))
	if err != nil {
		err = errors.New("error converting meeting image file: " + err.Error())
		return nil, err
	}

	if imageFile == nil {
		return nil, nil
	}

	var outputImage api.File
	if err := api.ConvertToOtherType(imageFile, &outputImage); err != nil {
		err = errors.New("error converting meeting image file to api.File: " + err.Error())
		return nil, err
	}
	outputImage.ID = imageFile.UUID
	return &outputImage, nil
}

func loadMeetingLocation(ctx context.Context, meeting Meeting) (api.Location, error) {
	location, err := meeting.GetLocation(Tx(ctx))
	if err != nil {
		err = errors.New("error converting meeting location: " + err.Error())
		return api.Location{}, err
	}

	return convertLocation(location), nil
}

func loadMeetingParticipants(ctx context.Context, meeting Meeting, user User) (api.MeetingParticipants, error) {
	tx := Tx(ctx)

	participants, err := meeting.Participants(tx, user)
	if err != nil {
		err = errors.New("error converting meeting participants: " + err.Error())
		return nil, err
	}

	outputParticipants, err := convertMeetingParticipants(ctx, participants)
	if err != nil {
		return nil, err
	}

	return outputParticipants, nil
}

// CreateInvites creates meeting invitations from a list of email addresses. The addresses
// can be comma-separated or newline separated.
func (m *Meeting) CreateInvites(ctx context.Context, emails string) error {
	if emails == "" {
		return nil
	}

	cUser := CurrentUser(ctx)
	tx := Tx(ctx)

	can, err := cUser.CanCreateMeetingInvite(tx, *m)
	if err != nil {
		return errors.New("error creating meeting invites, " + err.Error())
	}
	if !can {
		return errors.New("user cannot create invites for this meeting")
	}

	inv := MeetingInvite{
		MeetingID: m.ID,
		InviterID: cUser.ID,
	}

	badEmails := make([]string, 0)
	for _, email := range splitEmailList(emails) {
		inv.Email = email
		if err := inv.Create(tx); err != nil {
			badEmails = append(badEmails, email)
		}
	}
	if len(badEmails) > 0 {
		return fmt.Errorf("problems creating invitations, bad emails: %s", badEmails)
	}

	return nil
}

func splitEmailList(emails string) []string {
	if emails == "" {
		return []string{}
	}

	split := strings.Split(strings.ReplaceAll(strings.ReplaceAll(emails, "\r\n", ","), "\n", ","), ",")
	for i := range split {
		split[i] = strings.TrimSpace(split[i])
	}
	return split
}

// loadMeetingInvites gets the meeting's invites and converts into api.MeetingInvites
func loadMeetingInvites(tx *pop.Connection, meeting Meeting, user User) (api.MeetingInvites, error) {
	invites, err := meeting.Invites(tx, user)
	if err != nil {
		return api.MeetingInvites{}, err
	}
	output := convertMeetingInvites(meeting, invites)
	return output, nil
}

// convertMeetingInvites converts model.MeetingInvites into  api.MeetingInvites
func convertMeetingInvites(meeting Meeting, invites MeetingInvites) api.MeetingInvites {
	output := make(api.MeetingInvites, len(invites))
	for i := range output {
		output[i] = convertMeetingInvite(meeting, invites[i])
	}
	return output
}

func convertMeetingInvite(meeting Meeting, invite MeetingInvite) api.MeetingInvite {
	output := api.MeetingInvite{}
	output.MeetingID = meeting.UUID
	output.InviterID = invite.Inviter.UUID
	output.Email = invite.Email

	return output
}
