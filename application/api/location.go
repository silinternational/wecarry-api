package api

// Location gives the description of a Geographic location
//
// swagger:model
type Location struct {
	// Human-friendly description, limited to 255 characters, e.g. 'Los Angeles, CA, USA'
	Description string `json:"description"`

	// Country (ISO 3166-1 Alpha-2 code), e.g. 'US'
	Country string `json:"country"`

	// Latitude in decimal degrees, e.g. -30.95 = 30 degrees 57 minutes south
	Latitude float64 `json:"latitude"`

	// Longitude in decimal degrees, e.g. -80.05 = 80 degrees 3 minutes west
	Longitude float64 `json:"longitude"`
}
