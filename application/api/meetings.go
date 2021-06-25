package api

// MeetingName just gives the name of a meeting a/k/a Event, to serve as a focal point for finding, answering, carrying, and exchanging requests
// swagger:model
type MeetingName struct {
	// Short name, limited to 80 characters
	Name string `json:"name"`
}
