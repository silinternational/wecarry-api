package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/log"
)

// Watch is the model for storing request watches that trigger notifications on the conditions specified
type Watch struct {
	ID            int          `json:"-" db:"id"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
	UUID          uuid.UUID    `json:"uuid" db:"uuid"`
	OwnerID       int          `json:"owner_id" db:"owner_id"`
	Name          string       `json:"name" db:"name"`
	DestinationID nulls.Int    `json:"destination_id" db:"destination_id"`
	OriginID      nulls.Int    `json:"origin_id" db:"origin_id"`
	MeetingID     nulls.Int    `json:"meeting_id" db:"meeting_id"`
	SearchText    nulls.String `json:"search_text" db:"search_text"`
	Size          *RequestSize `json:"size" db:"size"`

	Destination *Location `belongs_to:"locations"`
	Origin      *Location `belongs_to:"locations"`
	Owner       User      `belongs_to:"users"`
	Meeting     *Meeting  `belongs_to:"meetings"`
}

// Watches is used for methods that operate on lists of objects
type Watches []Watch

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (w *Watch) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: w.UUID, Name: "UUID"},
		&validators.IntIsPresent{Field: w.OwnerID, Name: "OwnerID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (w *Watch) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (w *Watch) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create stores the Watch data as a new record in the database.
func (w *Watch) Create(tx *pop.Connection) error {
	return create(tx, w)
}

// Update writes the Watch data to an existing database record.
func (w *Watch) Update(tx *pop.Connection) error {
	return update(tx, w)
}

// FindByUUID loads from DB the Watch record identified by the given UUID
func (w *Watch) FindByUUID(tx *pop.Connection, id string) error {
	if id == "" {
		return errors.New("error: watch uuid must not be blank")
	}

	if err := tx.Where("uuid = ?", id).First(w); err != nil {
		return fmt.Errorf("error finding watch by uuid: %s", err.Error())
	}

	return nil
}

// DeleteForOwner deletes the watch with the given UUID, if the user is the owner
// of the watch.
func (w *Watch) DeleteForOwner(tx *pop.Connection, id string, user User) (string, *api.AppError) {
	if id == "" {
		appError := api.AppError{
			Category: api.CategoryUser,
			Key:      api.ErrorWatchMissingID,
			Err:      errors.New("error: watch uuid must not be blank"),
		}
		return "", &appError
	}

	if err := w.FindByUUID(tx, id); err != nil {
		appError := api.AppError{
			Category: api.CategoryNotFound,
			Key:      api.ErrorWatchNotFound,
			Err:      err,
		}
		return "", &appError
	}
	if w.OwnerID != user.ID {
		appError := api.AppError{
			Category: api.CategoryForbidden,
			Key:      api.ErrorNotAuthorized,
			Err:      errors.New("error: user may not delete a watch they don't own."),
		}
		return "", &appError
	}

	if err := w.Destroy(tx); err != nil {
		appError := api.AppError{
			Category: api.CategoryDatabase,
			Key:      api.ErrorWatchDeleteFailure,
			Err:      fmt.Errorf("error attempting to delete a watch with uuid %s: %s", w.UUID.String(), err.Error()),
		}
		return "", &appError
	}

	return w.UUID.String(), nil
}

// FindByUser returns all watches owned by the given user.
func (w *Watches) FindByUser(tx *pop.Connection, user User, eagerFields ...string) error {
	var q *pop.Query
	if len(eagerFields) > 0 {
		q = tx.EagerPreload(eagerFields...).Where("owner_id = ?", user.ID)
	} else {
		q = tx.Where("owner_id = ?", user.ID)
	}

	if err := q.Order("updated_at desc").All(w); err != nil {
		return fmt.Errorf("error getting watches for user id %v ... %v", user.ID, err)
	}

	return nil
}

// GetOwner returns the owner of the watch.
func (w *Watch) GetOwner(tx *pop.Connection) (*User, error) {
	owner := User{}
	if err := tx.Find(&owner, w.OwnerID); err != nil {
		return nil, err
	}
	return &owner, nil
}

// GetDestination does not check authorization
func (w *Watch) GetDestination(tx *pop.Connection) (*Location, error) {
	location := &Location{}
	if !w.DestinationID.Valid {
		return nil, nil
	}
	if err := tx.Find(location, w.DestinationID); err != nil {
		return nil, err
	}
	return location, nil
}

// GetOrigin does not check authorization
func (w *Watch) GetOrigin(tx *pop.Connection) (*Location, error) {
	location := &Location{}
	if !w.OriginID.Valid {
		return nil, nil
	}
	if err := tx.Find(location, w.OriginID); err != nil {
		return nil, err
	}
	return location, nil
}

// SetDestination sets the destination field, creating a new record in the database if necessary.
func (w *Watch) SetDestination(tx *pop.Connection, location Location) error {
	if w.DestinationID.Valid {
		location.ID = w.DestinationID.Int
		return location.Update(tx)
	}
	if err := location.Create(tx); err != nil {
		return err
	}
	w.DestinationID = nulls.NewInt(location.ID)
	return w.Update(tx)
}

// SetOrigin sets the origin field, creating a new record in the database if necessary.
func (w *Watch) SetOrigin(tx *pop.Connection, location Location) error {
	if w.OriginID.Valid {
		location.ID = w.OriginID.Int
		return location.Update(tx)
	}
	if err := location.Create(tx); err != nil {
		return err
	}
	w.OriginID = nulls.NewInt(location.ID)
	return w.Update(tx)
}

// Destroy wraps the Pop function of the same name
func (w *Watch) Destroy(tx *pop.Connection) error {
	return tx.Destroy(w)
}

func (w *Watch) LoadMeeting(tx *pop.Connection, user User) error {
	if w == nil || !w.MeetingID.Valid || user.ID != w.OwnerID {
		w.Meeting = nil
		return nil
	}
	meeting := &Meeting{}
	if err := tx.Find(meeting, w.MeetingID); err != nil {
		return err
	}
	w.Meeting = meeting
	return nil
}

// LoadForAPI assumes w.Destination and w.Origin were alread hydrated by an Eager query
func (w *Watch) LoadForAPI(tx *pop.Connection, user User) error {
	if user.ID != w.OwnerID {
		return errors.New("user is not watch owner")
	}

	return nil
}

// matchesRequest returns true if all non-null watch criteria match the request
func (w *Watch) matchesRequest(tx *pop.Connection, request Request) bool {
	if w == nil {
		log.Errorf("nil receiver in Watch.matchesRequest")
		return false
	}
	matchFunctions := []func(*Watch, *pop.Connection, Request) bool{
		(*Watch).sizeMatches,
		(*Watch).textMatches,
		(*Watch).meetingMatches,
		(*Watch).destinationMatches,
		(*Watch).originMatches,
	}
	for _, c := range matchFunctions {
		if !c(w, tx, request) {
			return false
		}
	}
	return true
}

// destinationMatches returns true if watch destination is not provided or passes the IsNear test
func (w *Watch) destinationMatches(tx *pop.Connection, request Request) bool {
	if w == nil {
		log.Errorf("nil receiver in Watch.destinationMatches")
		return false
	}
	if !w.DestinationID.Valid {
		return true
	}
	requestDestination, err := request.GetDestination(tx)
	if err != nil {
		log.Errorf("failed to get request %s destination in destinationMatches, %s", request.UUID, err)
		return false
	}
	watchDestination, err := w.GetDestination(tx)
	if err != nil {
		log.Errorf("failed to get watch %s destination in destinationMatches, %s", w.UUID, err)
	}

	if watchDestination == nil {
		return true
	}

	return watchDestination.IsNear(*requestDestination)
}

// originMatches returns true if watch origin is not provided or passes the IsNear test
func (w *Watch) originMatches(tx *pop.Connection, request Request) bool {
	if w == nil {
		log.Errorf("nil receiver in Watch.originMatches")
		return false
	}
	if !w.OriginID.Valid {
		return true
	}
	requestOrigin, err := request.GetOrigin(tx)
	if err != nil {
		log.Errorf("failed to get request %s origin in originMatches, %s", request.UUID, err)
		return false
	}
	if requestOrigin == nil {
		return false
	}

	watchOrigin, err := w.GetOrigin(tx)
	if err != nil {
		log.Errorf("failed to get watch %s origin in originMatches, %s", w.UUID, err)
	}
	if watchOrigin == nil {
		return true
	}
	return watchOrigin.IsNear(*requestOrigin)
}

// meetingMatches returns true if watch meeting is not provided or is identical to the request meeting
func (w *Watch) meetingMatches(tx *pop.Connection, request Request) bool {
	if w == nil {
		log.Errorf("nil receiver in Watch.meetingMatches")
		return false
	}
	if !w.MeetingID.Valid {
		return true
	}
	return w.MeetingID == request.MeetingID
}

// textMatches returns true if watch text is not provided or is in the request title, description or creator's nickname
func (w *Watch) textMatches(tx *pop.Connection, request Request) bool {
	if w == nil {
		log.Errorf("nil receiver in Watch.textMatches")
		return false
	}
	if !w.SearchText.Valid {
		return true
	}
	if strings.Contains(request.Title, w.SearchText.String) {
		return true
	}
	if strings.Contains(request.Description.String, w.SearchText.String) {
		return true
	}
	creator, err := request.Creator(tx)
	if err != nil {
		log.Errorf("failed to get request %s creator in textMatches, %s", request.UUID, err)
		return false
	}
	if strings.Contains(creator.Nickname, w.SearchText.String) {
		return true
	}
	return false
}

// sizeMatches returns true if watch size is larger or the same as the request size
func (w *Watch) sizeMatches(tx *pop.Connection, request Request) bool {
	if w == nil {
		log.Errorf("nil receiver in Watch.sizeMatches")
		return false
	}
	if w.Size == nil {
		return true
	}
	return w.Size.isLargerOrSame(request.Size)
}
