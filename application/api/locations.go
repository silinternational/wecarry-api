package api

// Location just gives the description of a Geographic location
// swagger:model
type LocationDescription struct {

	// Human-friendly description, limited to 255 characters, e.g. 'Los Angeles, CA, USA'
	Description string `json:"description"`
}
