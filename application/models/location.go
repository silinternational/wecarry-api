package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"

	"github.com/silinternational/wecarry-api/api"
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
func (l *Location) Create(tx *pop.Connection) error {
	return create(tx, l)
}

// Update writes the Location data to an existing database record.
func (l *Location) Update(tx *pop.Connection) error {
	return update(tx, l)
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
func (l *Locations) FindByIDs(tx *pop.Connection, ids []int) error {
	ids = domain.UniquifyIntSlice(ids)
	return tx.Where("id in (?)", ids).All(l)
}

// DeleteUnused removes all locations that are no longer used
func (l *Locations) DeleteUnused() error {
	if found, err := checkLocationsFks(); err != nil {
		return errors.New("error checking for new locations references, " + err.Error())
	} else if found {
		return errors.New("new key found, canceling location cleanup")
	}

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
		return fmt.Errorf("attempted to delete too many locations, unused=%d, MaxLocationDelete=%d",
			len(locations), domain.Env.MaxLocationDelete)
	}
	if len(locations) == 0 {
		return nil
	}

	nRemovedFromDB := 0
	for _, location := range locations {
		l := location
		if err := DB.Destroy(&l); err != nil {
			domain.ErrLogger.Printf("location %d destroy error, %s", location.ID, err)
			continue
		}
		nRemovedFromDB++
	}

	if nRemovedFromDB < len(locations) {
		domain.ErrLogger.Printf("not all unused locations were removed")
	}
	domain.Logger.Printf("removed %d from location table", nRemovedFromDB)
	return nil
}

func checkLocationsFks() (bool, error) {
	type KeyType struct {
		ForeignTable string `db:"foreign_table"`
		FkColumns    string `db:"fk_columns"`
	}
	var keys []KeyType
	if err := DB.RawQuery(`
SELECT kcu.table_name AS foreign_table, string_agg(kcu.column_name, ', ') as fk_columns
FROM information_schema.table_constraints tco
JOIN information_schema.key_column_usage kcu
          ON tco.constraint_schema = kcu.constraint_schema
          AND tco.constraint_name = kcu.constraint_name
JOIN information_schema.referential_constraints rco
          ON tco.constraint_schema = rco.constraint_schema
          AND tco.constraint_name = rco.constraint_name
JOIN information_schema.table_constraints rel_tco
          ON rco.unique_constraint_schema = rel_tco.constraint_schema
          AND rco.unique_constraint_name = rel_tco.constraint_name
WHERE tco.constraint_type = 'FOREIGN KEY' AND rel_tco.table_name = 'locations'
GROUP BY kcu.table_schema,
         kcu.table_name,
         rel_tco.table_name,
         rel_tco.table_schema,
         kcu.constraint_name
ORDER BY kcu.table_schema,
         kcu.table_name;`).All(&keys); err != nil {
		return false, err
	}
	if len(keys) != 6 {
		// expected 6 foreign keys: [{meetings location_id} {requests destination_id} {requests origin_id} {users location_id} {watches destination_id} {watches origin_id}]
		return true, nil
	}
	return false, nil
}

func convertLocation(location Location) api.Location {
	return api.Location{
		Description: location.Description,
		Country:     location.Country,
		Latitude:    location.Latitude,
		Longitude:   location.Longitude,
	}
}

func ConvertLocationInput(input api.LocationInput) Location {
	l := Location{
		Description: input.Description,
		Country:     input.Country,
	}

	domain.SetOptionalFloatField(input.Latitude, &l.Latitude)
	domain.SetOptionalFloatField(input.Longitude, &l.Longitude)

	return l
}
