package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// Trust is the model for storing Organization connections, also known as Trusts
type Trust struct {
	ID          int       `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	PrimaryID   int       `json:"primary_id" db:"primary_id"`
	SecondaryID int       `json:"secondary_id" db:"secondary_id"`
}

// Trusts is used for struct-attached functions that operate on lists of objects
type Trusts []Trust

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (t *Trust) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: t.PrimaryID, Name: "PrimaryID"},
		&validators.IntIsPresent{Field: t.PrimaryID, Name: "SecondaryID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (t *Trust) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (t *Trust) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create stores the Trust data as a new record in the database.
func (t *Trust) Create() error {
	return create(t)
}

// Update writes the Trust data to an existing database record.
func (t *Trust) Update() error {
	return update(t)
}

// FindByOrgIDs loads from DB the Trust record identified by the given Organization IDs. The two arguments are
// reversible.
func (t *Trust) FindByOrgIDs(id1, id2 int) error {
	if id1 <= 0 || id2 <= 0 {
		return errors.New("error: both organization IDs must be valid")
	}

	if err := DB.Where("primary_id = ? AND secondary_id = ? OR primary_id = ? AND secondary_id = ?",
		id1, id2, id2, id1).First(t); err != nil {
		return fmt.Errorf("error finding Trust by org ids, %s", err.Error())
	}

	return nil
}

// FindByOrgIDPrimary returns all trusts where a given Organization is the Primary org
func (t *Trusts) FindByOrgIDPrimary(id int) error {
	if err := DB.Where("primary_id = ?", id).All(t); err != nil {
		return err
	}

	return nil
}

// FindByOrgIDSecondary returns all trusts where a given Organization is the Secondary org
func (t *Trusts) FindByOrgIDSecondary(id int) error {
	if err := DB.Where("secondary_id = ?", id).All(t); err != nil {
		return err
	}

	return nil
}
