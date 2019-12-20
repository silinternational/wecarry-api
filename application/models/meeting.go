package models

import (
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/silinternational/wecarry-api/domain"
	"time"
)

// Meeting represents an event where people gather together from different locations
type Meeting struct {
	ID          int          `json:"id" db:"id"`
	UUID        uuid.UUID    `json:"uuid" db:"uuid"`
	Name        string       `json:"name" db:"name"`
	Description nulls.String `json:"description" db:"description"`
	MoreInfoURL nulls.String `json:"more_info_url" db:"more_info_url"`
	StartDate   time.Time    `json:"start_date" db:"start_date"`
	EndDate     time.Time    `json:"end_date" db:"end_date"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	CreatedByID int          `json:"created_by_id" db:"created_by_id"`
	ImageFileID nulls.Int    `json:"image_file_id" db:"image_file_id"`
	LocationID  int          `json:"location_id" db:"location_id"`

	CreatedBy User     `belongs_to:"users"`
	ImageFile File     `belongs_to:"files"`
	Location  Location `belongs_to:"locations"`
	Posts     Posts    `has_many:"posts" fk_id:"id" order_by:"updated_at desc"`
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

	v.Message = fmt.Sprintf("Start date must come no later thand end date chronologically. Got %s and %s.",
		v.StartDate.Format(domain.DateFormat), v.EndDate.Format(domain.DateFormat))
	errors.Add(validators.GenerateKey(v.Name), v.Message)
}

// FindByUUID finds a meeting by the UUID field and loads its CreatedBy field
func (m *Meeting) FindByUUID(uuid string) error {
	if uuid == "" {
		return errors.New("error finding meeting: uuid must not be blank")
	}

	if err := DB.Eager("CreatedBy").Where("uuid = ?", uuid).First(m); err != nil {
		return fmt.Errorf("error finding meeting by uuid: %s", err.Error())
	}

	return nil
}

func getOrdered(m *Meetings, q *pop.Query) error {
	return q.Order("start_date asc").All(m)
}

// FindOnDate finds the meetings that have StartDate before timeInFocus-date and an EndDate after it
// (inclusive on both)
func (m *Meetings) FindOnDate(timeInFocus time.Time) error {
	date := timeInFocus.Format(domain.DateTimeFormat)
	where := "start_date <= ? and end_date >= ?"

	if err := getOrdered(m, DB.Eager("CreatedBy").Where(where, date, date)); err != nil {
		return fmt.Errorf("error finding meeting with start_date and end_date straddling %s ... %s",
			date, err.Error())
	}

	return nil
}

// FindOnOrAfterDate finds the meetings that have an EndDate on or after the timeInFocus-date
func (m *Meetings) FindOnOrAfterDate(timeInFocus time.Time) error {

	date := timeInFocus.Format(domain.DateTimeFormat)

	if err := getOrdered(m, DB.Eager("CreatedBy").Where("end_date >= ?", date)); err != nil {
		return fmt.Errorf("error finding meeting with end_date before %s ... %s", date, err.Error())
	}

	return nil
}

// FindAfterDate finds the meetings that have a StartDate after the timeInFocus-date
func (m *Meetings) FindAfterDate(timeInFocus time.Time) error {
	date := timeInFocus.Format(domain.DateTimeFormat)

	if err := getOrdered(m, DB.Eager("CreatedBy").Where("start_date > ?", date)); err != nil {
		return fmt.Errorf("error finding meeting with start_date after %s ... %s", date, err.Error())
	}

	return nil
}

// FindRecent finds the meetings that have an EndDate within the past <domain.RecentMeetingDelay> days
// before timeInFocus-date (not inclusive)
func (m *Meetings) FindRecent(timeInFocus time.Time) error {
	yesterday := timeInFocus.Add(-domain.DurationDay).Format(domain.DateTimeFormat)
	recentDate := timeInFocus.Add(-domain.RecentMeetingDelay)
	where := "end_date between ? and ?"

	if err := getOrdered(m, DB.Eager("CreatedBy").Where(where, recentDate, yesterday)); err != nil {
		return fmt.Errorf("error finding meeting with end_date between %s and %s ... %s",
			recentDate, yesterday, err.Error())
	}

	return nil
}

// AttachImage assigns a previously-stored File to this Meeting as its image. Parameter `fileID` is the UUID
// of the image to attach.
func (m *Meeting) AttachImage(fileID string) (File, error) {
	var f File
	if err := f.FindByUUID(fileID); err != nil {
		err = fmt.Errorf("error finding meeting image with id %s ... %s", fileID, err)
		return f, err
	}

	m.ImageFileID = nulls.NewInt(f.ID)
	// if this is a new object, don't save it yet
	if m.ID != 0 {
		if err := DB.Update(m); err != nil {
			return f, err
		}
	}

	return f, nil
}

// GetImage retrieves the file attached as the Meeting Image
func (m *Meeting) GetImage() (*File, error) {
	if err := DB.Load(m, "ImageFile"); err != nil {
		return nil, err
	}

	if !m.ImageFileID.Valid {
		return nil, nil
	}

	if err := m.ImageFile.refreshURL(); err != nil {
		return nil, err
	}

	return &m.ImageFile, nil
}

func (m *Meeting) GetCreator() (*User, error) {
	creator := User{}
	if err := DB.Find(&creator, m.CreatedByID); err != nil {
		return nil, err
	}
	return &creator, nil
}

// GetLocation returns the related Location object.
func (m *Meeting) GetLocation() (Location, error) {
	location := Location{}
	if err := DB.Find(&location, m.LocationID); err != nil {
		return location, err
	}

	return location, nil
}

// SetLocation sets the location field, creating a new record in the database if necessary.
func (m *Meeting) SetLocation(location Location) error {
	location.ID = m.LocationID
	m.Location = location
	return DB.Update(&m.Location)
}

// Create stores the Meeting data as a new record in the database.
func (m *Meeting) Create() error {
	if m.UUID.Version() == 0 {
		m.UUID = domain.GetUUID()
	}

	valErrs, err := DB.ValidateAndCreate(m)
	if err != nil {
		return err
	}

	if len(valErrs.Errors) > 0 {
		vErrs := flattenPopErrors(valErrs)
		return errors.New(vErrs)
	}

	return nil
}

// Update writes the Meeting data to an existing database record.
func (m *Meeting) Update() error {
	valErrs, err := DB.ValidateAndUpdate(m)
	if err != nil {
		return err
	}

	if len(valErrs.Errors) > 0 {
		vErrs := flattenPopErrors(valErrs)
		return errors.New(vErrs)
	}

	return nil
}

// CanCreate returns a bool based on whether the current user is allowed to create a meeting
func (m *Meeting) CanCreate(user User) bool {
	return true
}

// CanUpdate returns a bool based on whether the current user is allowed to update a meeting
func (m *Meeting) CanUpdate(user User) bool {
	switch user.AdminRole {
	case UserAdminRoleSuperAdmin, UserAdminRoleSalesAdmin, UserAdminRoleAdmin:
		return true
	}

	return user.ID == m.CreatedByID
}
