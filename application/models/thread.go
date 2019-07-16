package models

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type Thread struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Uuid      uuid.UUID `json:"uuid" db:"uuid"`
	PostID    int       `json:"post_id" db:"post_id"`
	Post      Post      `belongs_to:"posts"`
}

// String is not required by pop and may be deleted
func (t Thread) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Threads is not required by pop and may be deleted
type Threads []Thread

// String is not required by pop and may be deleted
func (t Threads) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *Thread) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: t.ID, Name: "ID"},
		&validators.UUIDIsPresent{Field: t.Uuid, Name: "Uuid"},
		&validators.IntIsPresent{Field: t.PostID, Name: "PostID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *Thread) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *Thread) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
