package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	"github.com/silinternational/wecarry-api/domain"
)

type StandardPreferences struct {
	Language   string `json:"language"`
	TimeZone   string `json:"time_zone"`
	WeightUnit string `json:"weight_unit"`
}

type UserPreference struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UUID      uuid.UUID `json:"uuid" db:"uuid"`
	UserID    int       `json:"user_id" db:"user_id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	User      User      `belongs_to:"users"`
}

// String is not required by pop and may be deleted
func (s UserPreference) String() string {
	jm, _ := json.Marshal(s)
	return string(jm)
}

// UserPreferences is not required by pop and may be deleted
type UserPreferences []UserPreference

// String is not required by pop and may be deleted
func (p UserPreferences) String() string {
	jm, _ := json.Marshal(p)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (p *UserPreference) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: p.UUID, Name: "UUID"},
		&validators.IntIsPresent{Field: p.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: p.Key, Name: "Key"},
		&validators.StringIsPresent{Field: p.Value, Name: "Value"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (p *UserPreference) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (p *UserPreference) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// FindByUUID loads from DB the UserPreference record identified by the given UUID
func (p *UserPreference) FindByUUID(id string) error {
	if id == "" {
		return errors.New("error: user preference uuid must not be blank")
	}

	if err := DB.Where("uuid = ?", id).First(p); err != nil {
		return fmt.Errorf("error finding user preference by uuid: %s", err.Error())
	}

	return nil
}

// Save wraps DB.Save() call to create a UUID if it's empty and check for errors
func (p *UserPreference) Save() error {
	if p.UUID.Version() == 0 {
		p.UUID = domain.GetUUID()
	}

	validationErrs, err := p.Validate(DB)
	if validationErrs != nil && validationErrs.HasAny() {
		return errors.New(FlattenPopErrors(validationErrs))
	}
	if err != nil {
		return err
	}

	return DB.Save(p)
}
