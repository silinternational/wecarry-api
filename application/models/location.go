package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
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

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (l *Location) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: l.Description, Name: "Description"},
		&validators.StringLengthInRange{
			Name:    "country",
			Field:   l.Country,
			Min:     2,
			Max:     2,
			Message: "country must be ISO 3166–1 alpha-2",
		},
		&geoValidator{
			Name:      "geo",
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		},
	), nil
}

type geoValidator struct {
	Name, Message       string
	Latitude, Longitude nulls.Float64
}

// IsValid checks the latitude and longitude valid ranges
func (v *geoValidator) IsValid(errors *validate.Errors) {
	if !v.Latitude.Valid && !v.Longitude.Valid {
		return
	}

	if v.Latitude.Valid != v.Longitude.Valid {
		errors.Add(validators.GenerateKey(v.Name), "only one coordinate given, must have neither or both")
	}

	if v.Latitude.Float64 < -90.0 || v.Latitude.Float64 > 90.0 {
		v.Message = fmt.Sprintf("Latitude %v is out of range", v.Latitude)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}

	if v.Longitude.Float64 < -180.0 || v.Longitude.Float64 > 180.0 {
		v.Message = fmt.Sprintf("Longitude %v is out of range", v.Longitude)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}

	if v.Longitude.Float64 == 0 && v.Latitude.Float64 == 0 {
		v.Message = "a valid geo coordinate must be given"
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}
}

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

// Distance calculates the distance in km between two locations
func (l *Location) Distance(l2 Location) float64 {
	lat1 := l.Latitude.Float64
	lon1 := l.Longitude.Float64
	lat2 := l2.Latitude.Float64
	lon2 := l2.Longitude.Float64

	var p = math.Pi / 180

	var a = 0.5 - math.Cos((lat2-lat1)*p)/2 +
		math.Cos(lat1*p)*math.Cos(lat2*p)*
			(1-math.Cos((lon2-lon1)*p))/2

	return 12742 * math.Asin(math.Sqrt(a)) // 2 * R; R = 6371 km
}
