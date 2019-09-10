package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/validate/validators"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gofrs/uuid"
)

type Image struct {
	ID        int             `json:"id" db:"id"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
	UUID      uuid.UUID       `json:"uuid" db:"uuid"`
	PostID    int             `json:"post_id" db:"post_id"`
	Content   nulls.ByteSlice `json:"content" db:"content"`
	URL       nulls.String    `json:"url" db:"url"`
	Post      Post            `belongs_to:"posts"`
}

// String is not required by pop and may be deleted
func (i Image) String() string {
	ji, _ := json.Marshal(i)
	return string(ji)
}

// Images is not required by pop and may be deleted
type Images []Image

// String is not required by pop and may be deleted
func (i Images) String() string {
	ji, _ := json.Marshal(i)
	return string(ji)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (i *Image) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: i.UUID, Name: "UUID"},
		&validators.IntIsPresent{Field: i.PostID, Name: "ThreadID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (i *Image) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (i *Image) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
