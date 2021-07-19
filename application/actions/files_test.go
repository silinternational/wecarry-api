package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) verifyFile(expected models.File, actual api.File, msg string) {
	as.Equal(expected.UUID, actual.ID, msg+", ID is not correct")
	as.Equal(expected.URL, actual.URL, msg+", URL is not correct")
	as.True(expected.URLExpiration.Equal(actual.URLExpiration), msg+", URLExpiration is not correct")
	as.Equal(expected.Name, actual.Name, msg+", Name is not correct")
	as.Equal(expected.Size, actual.Size, msg+", Size is not correct")
	as.Equal(expected.ContentType, actual.ContentType, msg+", ContentType is not correct")
}
