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
	"github.com/silinternational/wecarry-api/domain"
)

type Location struct {
	ID          int           `json:"id" db:"id"`
	Description string        `json:"description" db:"description"`
	Country     string        `json:"country" db:"country"`
	Latitude    nulls.Float64 `json:"latitude" db:"latitude"`
	Longitude   nulls.Float64 `json:"longitude" db:"longitude"`
}

// String can be helpful for serializing the model
func (l Location) String() string {
	ji, _ := json.Marshal(l)
	return string(ji)
}

// Locations is merely for convenience and brevity
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
	return create(l)
}

// Update writes the Location data to an existing database record.
func (l *Location) Update() error {
	valErrs, err := DB.ValidateAndUpdate(l)
	if err != nil {
		return err
	}

	if len(valErrs.Errors) > 0 {
		vErrs := flattenPopErrors(valErrs)
		return errors.New(vErrs)
	}

	return nil
}

// DistanceKm calculates the distance in km between two locations
func (l *Location) DistanceKm(loc2 Location) float64 {
	if !l.Latitude.Valid || !l.Longitude.Valid || !loc2.Latitude.Valid || !loc2.Longitude.Valid {
		return math.NaN()
	}

	lat1 := l.Latitude.Float64
	lon1 := l.Longitude.Float64
	lat2 := loc2.Latitude.Float64
	lon2 := loc2.Longitude.Float64

	// Haversine formula implementation derived from Stack Overflow answer:
	// https://stackoverflow.com/a/21623206
	var p = math.Pi / 180

	var a = 0.5 - math.Cos((lat2-lat1)*p)/2 +
		math.Cos(lat1*p)*math.Cos(lat2*p)*
			(1-math.Cos((lon2-lon1)*p))/2

	return 12742 * math.Asin(math.Sqrt(a)) // 2 * R; R = 6371 km
}

// IsNear answers the question "Are these two locations near each other?"
func (l *Location) IsNear(loc2 Location) bool {
	if l.Country != "" && l.Country == loc2.Country {
		return true
	}

	if d := l.DistanceKm(loc2); !math.IsNaN(d) && d < domain.DefaultProximityDistanceKm {
		return true
	}

	return false
}
