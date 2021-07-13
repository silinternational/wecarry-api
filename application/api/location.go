package api

import "github.com/gobuffalo/nulls"

// Location gives the description of a Geographic location
//
// swagger:model
type Location struct {
	// Human-friendly description, limited to 255 characters, e.g. 'Los Angeles, CA, USA'
	Description string `json:"description"`

	// Country (ISO 3166-1 Alpha-2 code), e.g. 'US'
	Country string `json:"country"`

	// Equivalent to Google Place's administrative area level 1
	State string `json:"state"`

	// Equivalent to Google Place's administrative area level 2
	County string `json:"county"`

	// Equivalent to Google Place's locality
	City string `json:"city"`

	// Equivalent to Google Place's sub-locality
	Borough string `json:"borough"`

	// Latitude in decimal degrees, e.g. -30.95 = 30 degrees 57 minutes south
	Latitude nulls.Float64 `json:"latitude"`

	// Longitude in decimal degrees, e.g. -80.05 = 80 degrees 3 minutes west
	Longitude nulls.Float64 `json:"longitude"`
}
