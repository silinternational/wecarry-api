package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/silinternational/wecarry-api/domain"
)

type StandardPreferences struct {
	Language   string `json:"language"`
	TimeZone   string `json:"time_zone"`
	WeightUnit string `json:"weight_unit"`
}

func (s *StandardPreferences) hydrateValues(values map[string]string) {
	s.Language = values[domain.UserPreferenceKeyLanguage]
	s.TimeZone = values[domain.UserPreferenceKeyTimeZone]
	s.WeightUnit = values[domain.UserPreferenceKeyWeightUnit]
}

type UserPreference struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UUID      uuid.UUID `json:"uuid" db:"uuid"`
	UserID    int       `json:"user_id" db:"user_id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
}

// String can be helpful for serializing the model
func (p UserPreference) String() string {
	jm, _ := json.Marshal(p)
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
func (p *UserPreference) FindByUUID(tx *pop.Connection, id string) error {
	if id == "" {
		return errors.New("error: user preference uuid must not be blank")
	}

	if err := tx.Where("uuid = ?", id).First(p); err != nil {
		return fmt.Errorf("error finding user preference by uuid: %s", err.Error())
	}

	return nil
}

// Save wraps tx.Save() call to create a UUID if it's empty and check for errors
func (p *UserPreference) Save(tx *pop.Connection) error {
	return save(tx, p)
}

type fieldAndValidator struct {
	fieldValue string
	validator  func(string) bool
}

func getPreferencesFieldsAndValidators(prefs StandardPreferences) map[string]fieldAndValidator {
	fieldAndValidators := map[string]fieldAndValidator{}
	fieldAndValidators[domain.UserPreferenceKeyLanguage] = fieldAndValidator{
		fieldValue: prefs.Language,
		validator:  domain.IsLanguageAllowed,
	}
	fieldAndValidators[domain.UserPreferenceKeyTimeZone] = fieldAndValidator{
		fieldValue: prefs.TimeZone,
		validator:  domain.IsTimeZoneAllowed,
	}
	fieldAndValidators[domain.UserPreferenceKeyWeightUnit] = fieldAndValidator{
		fieldValue: prefs.WeightUnit,
		validator:  domain.IsWeightUnitAllowed,
	}

	return fieldAndValidators
}

func (p *UserPreference) createForUser(tx *pop.Connection, user User, key, value string) error {
	if user.ID <= 0 {
		return errors.New("invalid user ID in userpreference.createForUser")
	}

	_ = tx.Where("user_id = ?", user.ID).Where("key = ?", key).First(p)
	if p.ID > 0 {
		err := fmt.Errorf("can't create UserPreference with key %s.  Already exists with id %v.", key, p.ID)
		return err
	}

	p.UserID = user.ID
	p.Key = key
	p.Value = value

	return p.Save(tx)
}

func (p *UserPreference) getForUser(tx *pop.Connection, user User, key string) error {
	err := tx.Where("user_id = ?", user.ID).Where("key = ?", key).First(p)
	if domain.IsOtherThanNoRows(err) {
		return err
	}

	return nil
}

// updateForUserByKey will also create a new instance, if a match is not found for that user
func (p *UserPreference) updateForUserByKey(tx *pop.Connection, user User, key, value string) error {
	err := p.getForUser(tx, user, key)
	if err != nil {
		return err
	}

	if p.ID == 0 {
		return p.createForUser(tx, user, key, value)
	}

	if p.Value == value {
		return nil
	}

	p.Value = value

	return p.Save(tx)
}

func updateUsersStandardPreferences(tx *pop.Connection, user User, prefs StandardPreferences) error {
	fieldAndValidators := getPreferencesFieldsAndValidators(prefs)
	for fieldName, fV := range fieldAndValidators {
		var p UserPreference

		if fV.fieldValue == "" {
			if err := p.remove(tx, user, fieldName); err != nil {
				return fmt.Errorf("error removing preference %s, %s", fieldName, err)
			}
			continue
		}

		if !fV.validator(fV.fieldValue) {
			return fmt.Errorf("unexpected UserPreference %s ... %s", fieldName, fV.fieldValue)
		}

		err := p.updateForUserByKey(tx, user, fieldName, fV.fieldValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *UserPreference) removeAll(tx *pop.Connection, userID int) error {
	return tx.RawQuery("DELETE FROM user_preferences WHERE user_id = ?", userID).Exec()
}

func (p *UserPreference) remove(tx *pop.Connection, user User, key string) error {
	if err := p.getForUser(tx, user, key); err != nil {
		return err
	}
	return tx.Destroy(p)
}
