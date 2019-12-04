package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
)

type PostHistory struct {
	ID         int        `json:"id" db:"id"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	Status     PostStatus `json:"status" db:"status"`
	PostID     int        `json:"post_id" db:"post_id"`
	ReceiverID nulls.Int  `json:"receiver_id" db:"receiver_id"`
	ProviderID nulls.Int  `json:"provider_id" db:"provider_id"`
	Post       Post       `belongs_to:"posts"`
	Receiver   User       `belongs_to:"users"`
	Provider   User       `belongs_to:"users"`
}

// String is not required by pop and may be deleted
func (p PostHistory) String() string {
	jt, _ := json.Marshal(p)
	return string(jt)
}

// PostHistories is not required by pop and may be deleted
type PostHistories []PostHistory

// String is not required by pop and may be deleted
func (p PostHistories) String() string {
	jt, _ := json.Marshal(p)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (p *PostHistory) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (p *PostHistory) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (p *PostHistory) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Load reads the selected fields from the database
func (p *PostHistory) Load(fields ...string) error {
	if err := DB.Load(p, fields...); err != nil {
		return fmt.Errorf("error loading data for post history %v, %s", p.ID, err)
	}

	return nil
}
