package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
)

// Watch is the model for storing post watches that trigger notifications on the conditions specified
type Watch struct {
	ID         int       `json:"id" db:"id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	UUID       uuid.UUID `json:"uuid" db:"uuid"`
	OwnerID    int       `json:"owner_id" db:"owner_id"`
	LocationID nulls.Int `json:"location_id" db:"location_id"`
}

// String can be helpful for serializing the model
func (w Watch) String() string {
	jw, _ := json.Marshal(w)
	return string(jw)
}

// Watches is merely for convenience and brevity
type Watches []Watch

// String can be helpful for serializing the model
func (w Watches) String() string {
	jw, _ := json.Marshal(w)
	return string(jw)
}

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
func (w *Watch) Create() error {
	return create(w)
}

// Update writes the Watch data to an existing database record.
func (w *Watch) Update() error {
	return update(w)
}

// FindByUUID loads from DB the Watch record identified by the given UUID
func (w *Watch) FindByUUID(id string) error {
	if id == "" {
		return errors.New("error: watch uuid must not be blank")
	}

	if err := DB.Where("uuid = ?", id).First(w); err != nil {
		return fmt.Errorf("error finding watch by uuid: %s", err.Error())
	}

	return nil
}

// FindByUser returns all watches owned by the given user.
func (w *Watches) FindByUser(user User) error {
	if err := DB.Where("owner_id = ?", user.ID).Order("updated_at desc").All(w); err != nil {
		return err
	}

	return nil
}

// GetOwner returns the owner of the watch.
func (w *Watch) GetOwner() (*User, error) {
	owner := User{}
	if err := DB.Find(&owner, w.OwnerID); err != nil {
		return nil, err
	}
	return &owner, nil
}

// GetLocation returns the related Location object.
func (w *Watch) GetLocation() (*Location, error) {
	location := &Location{}
	if !w.LocationID.Valid {
		return nil, nil
	}
	if err := DB.Find(location, w.LocationID); err != nil {
		return nil, err
	}
	return location, nil
}

// SetLocation sets the location field, creating a new record in the database if necessary.
func (w *Watch) SetLocation(location Location) error {
	if w.LocationID.Valid {
		location.ID = w.LocationID.Int
		return location.Update()
	}
	if err := location.Create(); err != nil {
		return err
	}
	w.LocationID = nulls.NewInt(location.ID)
	return w.Update()
}
