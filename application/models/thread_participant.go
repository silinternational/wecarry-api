package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type ThreadParticipant struct {
	ID           int       `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	ThreadID     int       `json:"thread_id" db:"thread_id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Thread       Thread    `belongs_to:"threads"`
	User         User      `belongs_to:"users"`
	LastViewedAt time.Time `json:"last_viewed_at" db:"last_viewed_at"`
}

// String is not required by pop and may be deleted
func (t ThreadParticipant) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// ThreadParticipants is not required by pop and may be deleted
type ThreadParticipants []ThreadParticipant

// String is not required by pop and may be deleted
func (t ThreadParticipants) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *ThreadParticipant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: t.ThreadID, Name: "ThreadID"},
		&validators.IntIsPresent{Field: t.UserID, Name: "UserID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *ThreadParticipant) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *ThreadParticipant) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// SetLastViewedAt sets the last viewed time field
func (t *ThreadParticipant) SetLastViewedAt(time time.Time) error {
	t.LastViewedAt = time
	return DB.Update(t)
}
