package models

import (
	"encoding/json"
	"errors"

	"github.com/gobuffalo/nulls"
)

type Location struct {
	ID          int           `json:"id" db:"id"`
	Description string        `json:"description" db:"description"`
	Country     string        `json:"country" db:"country"`
	Latitude    nulls.Float64 `json:"latitude" db:"latitude"`
	Longitude   nulls.Float64 `json:"longitude" db:"longitude"`
}

// String is not required by pop and may be deleted
func (l Location) String() string {
	ji, _ := json.Marshal(l)
	return string(ji)
}

// Locations is not required by pop and may be deleted
type Locations []Location

// Create stores the Location data as a new record in the database.
func (l *Location) Create() error {
	valErrs, err := DB.ValidateAndCreate(l)
	if err != nil {
		return err
	}

	if len(valErrs.Errors) > 0 {
		vErrs := FlattenPopErrors(valErrs)
		return errors.New(vErrs)
	}

	return nil
}
