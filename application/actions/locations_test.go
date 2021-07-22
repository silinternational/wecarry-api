package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) verifyLocation(expected models.Location, actual api.Location, msg string) {
	as.Equal(expected.Description, actual.Description, msg+", Description is not correct")
	as.Equal(expected.Country, actual.Country, msg+", Country is not correct")
	as.Equal(expected.State, actual.State, msg+", State is not correct")
	as.Equal(expected.County, actual.County, msg+", County is not correct")
	as.Equal(expected.City, actual.City, msg+", City is not correct")
	as.Equal(expected.Borough, actual.Borough, msg+", Borough is not correct")
	as.Equal(expected.Latitude, actual.Latitude, msg+", Latitude is not correct")
	as.Equal(expected.Longitude, actual.Longitude, msg+", Longitude is not correct")
}
