package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
)

type RequestFile struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	RequestID int       `json:"request_id" db:"request_id"`
	FileID    int       `json:"file_id" db:"file_id"`
	File      File      `belongs_to:"files"`
}

// String can be helpful for serializing the model
func (p RequestFile) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// RequestFiles is merely for convenience and brevity
type RequestFiles []RequestFile

// String can be helpful for serializing the model
func (p RequestFiles) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *RequestFile) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (p *RequestFile) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (p *RequestFile) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create stores the RequestFile data as a new record in the database.
func (p *RequestFile) Create() error {
	return create(p)
}
