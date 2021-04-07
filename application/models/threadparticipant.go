package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
)

type ThreadParticipant struct {
	ID             int       `json:"id" db:"id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	ThreadID       int       `json:"thread_id" db:"thread_id"`
	UserID         int       `json:"user_id" db:"user_id"`
	LastViewedAt   time.Time `json:"last_viewed_at" db:"last_viewed_at"`
	LastNotifiedAt time.Time `json:"last_notified_at" db:"last_notified_at"`
	Thread         Thread    `belongs_to:"threads"`
}

// String can be helpful for serializing the model
func (t ThreadParticipant) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// ThreadParticipants is merely for convenience and brevity
type ThreadParticipants []ThreadParticipant

// String can be helpful for serializing the model
func (t ThreadParticipants) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (t *ThreadParticipant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: t.ThreadID, Name: "ThreadID"},
		&validators.IntIsPresent{Field: t.UserID, Name: "UserID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (t *ThreadParticipant) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (t *ThreadParticipant) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// UpdateLastViewedAt sets the last viewed time field and writes to the database
func (t *ThreadParticipant) UpdateLastViewedAt(lastViewedAt time.Time) error {
	t.LastViewedAt = lastViewedAt
	if err := t.Update(); err != nil {
		return fmt.Errorf("failed to update thread_participant.last_viewed_at, %s", err)
	}
	return nil
}

// FindByThreadIDAndUserID reads a record by the given Thread ID and User ID
func (t *ThreadParticipant) FindByThreadIDAndUserID(threadID, userID int) error {
	if threadID <= 0 || userID <= 0 {
		return fmt.Errorf("error finding thread_participant, invalid id ... threadID %v, userID %v", threadID, userID)
	}

	if err := DB.Where("user_id = ? AND thread_id = ?", userID, threadID).First(t); err != nil {
		return fmt.Errorf("failed to find thread_participant record for user %d and thread %d, %s",
			userID, threadID, err)
	}
	return nil
}

// UpdateLastNotifiedAt sets LastNotifiedAt and writes to the database
func (t *ThreadParticipant) UpdateLastNotifiedAt(newTime time.Time) error {
	t.LastNotifiedAt = newTime
	if err := t.Update(); err != nil {
		return fmt.Errorf("failed to update thread_participant.last_notified_at, %s", err)
	}
	return nil
}

// Create stores the ThreadParticipant data as a new record in the database.
func (t *ThreadParticipant) Create() error {
	return create(t)
}

// Update writes the ThreadParticipant data to an existing database record.
func (t *ThreadParticipant) Update() error {
	return update(t)
}
