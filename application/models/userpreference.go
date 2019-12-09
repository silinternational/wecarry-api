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

func (s *StandardPreferences) hydrateValues(values [3]string) {
	s.Language = values[0]
	s.TimeZone = values[1]
	s.WeightUnit = values[2]

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

// String can be helpful for serializing the model
func (s UserPreference) String() string {
	jm, _ := json.Marshal(s)
	return string(jm)
}

// UserPreferences is merely for convenience and brevity
type UserPreferences []UserPreference

// String can be helpful for serializing the model
func (p UserPreferences) String() string {
	jm, _ := json.Marshal(p)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *UserPreference) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: p.UUID, Name: "UUID"},
		&validators.IntIsPresent{Field: p.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: p.Key, Name: "Key"},
		&validators.StringIsPresent{Field: p.Value, Name: "Value"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (p *UserPreference) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
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

func getPreferencesFieldsAndValidators(prefs StandardPreferences) ([3]string, [3]string, [3]func(string) bool) {
	fieldNames := [3]string{
		domain.UserPreferenceKeyLanguage,
		domain.UserPreferenceKeyTimeZone,
		domain.UserPreferenceKeyWeightUnit,
	}
	fields := [3]string{prefs.Language, prefs.TimeZone, prefs.WeightUnit}
	validators := [3]func(string) bool{domain.IsLanguageAllowed, domain.IsTimeZoneAllowed, domain.IsWeightUnitAllowed}
	return fieldNames, fields, validators
}

func (p *UserPreference) createForUser(user User, key, value string) error {

	if user.ID <= 0 {
		return errors.New("invalid user ID in userpreference.createForUser.")
	}

	_ = DB.Where("user_id = ?", user.ID).Where("key = ?", key).First(p)
	if p.ID > 0 {
		err := fmt.Errorf("can't create UserPreference with key %s.  Already exists with id %v.", key, p.ID)
		return err
	}

	p.UserID = user.ID
	p.Key = key
	p.Value = value

	return p.Save()
}

func (p *UserPreference) getForUser(user User, key string) error {
	err := DB.Where("user_id = ?", user.ID).Where("key = ?", key).First(p)
	if domain.IsOtherThanNoRows(err) {
		return err
	}

	return nil
}

// updateForUserByKey will also create a new instance, if a match is not found for that user
func (p *UserPreference) updateForUserByKey(user User, key, value string) error {

	err := p.getForUser(user, key)
	if err != nil {
		return err
	}

	if p.ID == 0 {
		return p.createForUser(user, key, value)
	}

	if p.Value == value {
		return nil
	}

	p.Value = value

	return p.Save()
}
