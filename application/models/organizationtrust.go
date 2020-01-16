package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"

	"github.com/silinternational/wecarry-api/domain"
)

// OrganizationTrust is the model for storing Organization connections, also known as OrganizationTrusts
type OrganizationTrust struct {
	ID          int       `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	PrimaryID   int       `json:"primary_id" db:"primary_id"`
	SecondaryID int       `json:"secondary_id" db:"secondary_id"`
}

// OrganizationTrusts is used for struct-attached functions that operate on lists of objects
type OrganizationTrusts []OrganizationTrust

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (o *OrganizationTrust) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: o.PrimaryID, Name: "PrimaryID"},
		&validators.IntIsPresent{Field: o.SecondaryID, Name: "SecondaryID"},
		&validators.IntsAreNotEqual{ValueOne: o.PrimaryID, ValueTwo: o.SecondaryID, Name: "SecondaryEqualsPrimary"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (o *OrganizationTrust) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (o *OrganizationTrust) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create stores the OrganizationTrust data as a new record in the database. Also creates a mirrored copy of the given OrganizationTrust.
func (o *OrganizationTrust) Create() error {
	var t2 OrganizationTrust
	if err := t2.FindByOrgIDs(o.PrimaryID, o.SecondaryID); err == nil {
		// already exists
		return nil
	} else if domain.IsOtherThanNoRows(err) {
		return err
	}
	if err := create(o); err != nil {
		return err
	}
	mirror := OrganizationTrust{
		PrimaryID:   o.SecondaryID,
		SecondaryID: o.PrimaryID,
	}
	if err := create(&mirror); err != nil {
		_ = DB.Destroy(o)
		return err
	}
	return nil
}

// Remove destroys two OrganizationTrust records identified by the given Organization IDs
func (o *OrganizationTrust) Remove(orgID1, orgID2 int) error {
	var t1, t2 OrganizationTrust
	if err := t1.FindByOrgIDs(orgID1, orgID2); err != nil {
		return fmt.Errorf("remove OrganizationTrust failed locating the first OrganizationTrust record, %s", err)
	}
	if err := t2.FindByOrgIDs(orgID2, orgID1); err != nil {
		return fmt.Errorf("remove OrganizationTrust failed locating the second OrganizationTrust record, %s", err)
	}
	if err := DB.Destroy(&t1); err != nil {
		return fmt.Errorf("remove OrganizationTrust failed to destroy the first OrganizationTrust record, %s", err)
	}
	return DB.Destroy(&t2)
}

// FindByOrgIDs loads from DB the OrganizationTrust record identified by the given Organization IDs.
func (o *OrganizationTrust) FindByOrgIDs(id1, id2 int) error {
	if id1 <= 0 || id2 <= 0 {
		return errors.New("error: both organization IDs must be valid")
	}

	if err := DB.Where("primary_id = ? AND secondary_id = ?", id1, id2).First(o); err != nil {
		return fmt.Errorf("error finding OrganizationTrust by org ids, %s", err.Error())
	}

	return nil
}

// FindByOrgID returns all trusts where a given Organization is the Primary org
func (o *OrganizationTrusts) FindByOrgID(id int) error {
	if err := DB.Where("primary_id = ?", id).All(o); err != nil {
		return err
	}

	return nil
}
