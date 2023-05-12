package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
)

// OrganizationTrust is the model for storing Organization connections, also known as OrganizationTrusts
type OrganizationTrust struct {
	ID          int       `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	PrimaryID   int       `json:"primary_id" db:"primary_id"`
	SecondaryID int       `json:"secondary_id" db:"secondary_id"`
}

// OrganizationTrusts is used for methods that operate on lists of objects
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

// CreateSymmetric creates two records to make a two-way trust connection
func (o *OrganizationTrust) CreateSymmetric(tx *pop.Connection) error {
	if err := o.Create(tx); err != nil {
		return err
	}

	mirror := OrganizationTrust{
		PrimaryID:   o.SecondaryID,
		SecondaryID: o.PrimaryID,
	}
	if err := mirror.Create(tx); err != nil {
		msg := fmt.Sprintf("failed to create mirror of trust %d - %d, %s", o.PrimaryID, o.SecondaryID, err)
		err = tx.Destroy(o)
		if err != nil {
			return fmt.Errorf("failed to delete primary trust after %s, %s", msg, err)
		}
		return errors.New(msg)
	}

	return nil
}

// Create stores an OrganizationTrust record with the given Organization IDs
func (o *OrganizationTrust) Create(tx *pop.Connection) error {
	trust := *o
	if err := trust.FindByOrgIDs(tx, o.PrimaryID, o.SecondaryID); err == nil {
		return nil
	} else if domain.IsOtherThanNoRows(err) {
		return err
	}
	return create(tx, &trust)
}

// RemoveSymmetric destroys two OrganizationTrust records identified by the given Organization IDs
func (o *OrganizationTrust) RemoveSymmetric(tx *pop.Connection, orgID1, orgID2 int) error {
	err1 := o.Remove(tx, orgID1, orgID2)
	err2 := o.Remove(tx, orgID2, orgID1)
	if err1 != nil {
		return fmt.Errorf("remove Trust failed on the first record, %s", err1)
	}
	if err2 != nil {
		return fmt.Errorf("remove Trust failed on the second record, %s", err2)
	}
	return nil
}

// Remove destroys the OrganizationTrust record identified by the given Organization IDs
func (o *OrganizationTrust) Remove(tx *pop.Connection, orgID1, orgID2 int) error {
	var trust OrganizationTrust
	if err := trust.FindByOrgIDs(tx, orgID1, orgID2); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return err
		}
		log.Errorf("no record found when removing organization trust %d - %d", orgID1, orgID2)
		return nil
	}
	return tx.Destroy(&trust)
}

// FindByOrgIDs loads from DB the OrganizationTrust record identified by the given Organization IDs.
func (o *OrganizationTrust) FindByOrgIDs(tx *pop.Connection, id1, id2 int) error {
	if id1 <= 0 || id2 <= 0 {
		return errors.New("error: both organization IDs must be valid")
	}

	if err := tx.Where("primary_id = ? AND secondary_id = ?", id1, id2).First(o); err != nil {
		return fmt.Errorf("error finding OrganizationTrust by org ids, %s", err.Error())
	}

	return nil
}

// FindByOrgID returns all trusts where a given Organization is the Primary org
func (o *OrganizationTrusts) FindByOrgID(tx *pop.Connection, id int) error {
	if err := tx.Where("primary_id = ?", id).All(o); err != nil {
		return err
	}

	return nil
}
