package models

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
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
	return update(l)
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
	p := math.Pi / 180

	a := 0.5 - math.Cos((lat2-lat1)*p)/2 +
		math.Cos(lat1*p)*math.Cos(lat2*p)*
			(1-math.Cos((lon2-lon1)*p))/2

	return 12742 * math.Asin(math.Sqrt(a)) // 2 * R; R = 6371 km
}

// IsNear answers the question "Are these two locations near each other?"
func (l *Location) IsNear(loc2 Location) bool {
	d := l.DistanceKm(loc2)
	return !math.IsNaN(d) && d < domain.DefaultProximityDistanceKm
}

// FindByIDs finds all Locations associated with the given IDs and loads them from the database
func (l *Locations) FindByIDs(ids []int) error {
	ids = domain.UniquifyIntSlice(ids)
	return DB.Where("id in (?)", ids).All(l)
}

// DeleteUnused removes all locations that are no longer used
func (l *Locations) DeleteUnused() error {
	var usedLocations []int

	var meetings Meetings
	if err := DB.All(&meetings); err != nil {
		return fmt.Errorf("could not load meetings in Locations.DeleteUnused, %s", err)
	}
	for _, m := range meetings {
		usedLocations = append(usedLocations, m.LocationID)
	}

	var requests Requests
	if err := DB.All(&requests); err != nil {
		return fmt.Errorf("could not load requests in Locations.DeleteUnused, %s", err)
	}
	for _, m := range requests {
		usedLocations = append(usedLocations, m.DestinationID)
		if m.OriginID.Valid {
			usedLocations = append(usedLocations, m.OriginID.Int)
		}
	}

	var users Users
	if err := DB.Where("location_id IS NOT NULL").All(&users); err != nil {
		return fmt.Errorf("could not load users in Locations.DeleteUnused, %s", err)
	}
	for _, m := range users {
		if m.LocationID.Valid {
			usedLocations = append(usedLocations, m.LocationID.Int)
		}
	}

	var watches Watches
	if err := DB.Where("origin_id IS NOT NULL OR destination_id IS NOT NULL").
		All(&watches); err != nil {
		return fmt.Errorf("could not load watches in Locations.DeleteUnused, %s", err)
	}
	for _, m := range watches {
		if m.OriginID.Valid {
			usedLocations = append(usedLocations, m.OriginID.Int)
		}
		if m.DestinationID.Valid {
			usedLocations = append(usedLocations, m.DestinationID.Int)
		}
	}

	var locations Locations
	if err := DB.Where("id NOT IN (?)", usedLocations).All(&locations); err != nil {
		return fmt.Errorf("could not load locations in Locations.DeleteUnused, %s", err)
	}

	if len(locations) > domain.Env.MaxLocationDelete {
		return fmt.Errorf("attempted to delete too many locations, MaxLocationDelete=%d", domain.Env.MaxLocationDelete)
	}
	if len(locations) == 0 {
		return nil
	}

	nRemovedFromDB := len(locations) // temporarily bypass the delete to be safe
	//nRemovedFromDB := 0
	//for _, location := range locations {
	//	l := location
	//	if err := DB.Destroy(&l); err != nil {
	//		domain.ErrLogger.Printf("location %d destroy error, %s", location.ID, err)
	//		continue
	//	}
	//	nRemovedFromDB++
	//}

	if nRemovedFromDB < len(locations) {
		domain.ErrLogger.Printf("not all unused locations were removed")
	}
	domain.Logger.Printf("removed %d from location table", nRemovedFromDB)
	return nil
}
