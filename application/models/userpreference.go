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
)

type UserSetting struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Uuid      uuid.UUID `json:"uuid" db:"uuid"`
	UserID    int       `json:"user_id" db:"user_id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	User      User      `belongs_to:"users"`
}

// String is not required by pop and may be deleted
func (s UserSetting) String() string {
	jm, _ := json.Marshal(s)
	return string(jm)
}

// UserSettings is not required by pop and may be deleted
type UserSettings []UserSetting

// String is not required by pop and may be deleted
func (s UserSettings) String() string {
	jm, _ := json.Marshal(s)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *UserSetting) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: s.Uuid, Name: "Uuid"},
		&validators.IntIsPresent{Field: s.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: s.Key, Name: "Key"},
		&validators.StringIsPresent{Field: s.Value, Name: "Value"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (s *UserSetting) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (s *UserSetting) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// FindByUUID loads from DB the UserSetting record identified by the given UUID
func (s *UserSetting) FindByUUID(id string, selectFields ...string) error {
	if id == "" {
		return errors.New("error: user setting uuid must not be blank")
	}

	if err := DB.Where("uuid = ?", id).Select(selectFields...).First(s); err != nil {
		return fmt.Errorf("error finding user setting by uuid: %s", err.Error())
	}

	return nil
}
