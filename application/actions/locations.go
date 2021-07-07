package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func convertLocation(location models.Location) api.Location {
	return api.Location{
		Description: location.Description,
		Country:     location.Country,
		Latitude:    location.Latitude,
		Longitude:   location.Longitude,
	}
}

func convertLocationInput(input api.LocationInput) models.Location {
	l := models.Location{
		Description: input.Description,
		Country:     input.Country,
	}

	domain.SetOptionalFloatField(input.Latitude, &l.Latitude)
	domain.SetOptionalFloatField(input.Longitude, &l.Longitude)

	return l
}
