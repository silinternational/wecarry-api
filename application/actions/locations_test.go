package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) verifyLocation(expected models.Location, actual api.Location, msg string) {
	as.Equal(expected.Description, actual.Description, msg+", Description is not correct")
	as.Equal(expected.Country, actual.Country, msg+", Country is not correct")
	as.Equal(expected.Latitude, actual.Latitude, msg+", Latitude is not correct")
	as.Equal(expected.Longitude, actual.Longitude, msg+", Longitude is not correct")
}
