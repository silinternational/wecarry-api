package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) verifyMeeting(expected models.Meeting, actual api.Meeting, msg string) {
	as.Equal(expected.UUID, actual.ID, msg+", ID is not correct")
	as.Equal(expected.Name, actual.Name, msg+", Name is not correct")
	as.Equal(expected.Description.String, actual.Description, msg+", Description is not correct")
	as.True(expected.StartDate.Equal(actual.StartDate), msg+", StartDate is not correct")
	as.True(expected.EndDate.Equal(actual.EndDate), msg+", EndDate is not correct")
	as.True(expected.CreatedAt.Equal(actual.CreatedAt), msg+", CreatedAt is not correct")
	as.True(expected.UpdatedAt.Equal(actual.UpdatedAt), msg+", UpdatedAt is not correct")
	as.Equal(expected.MoreInfoURL.String, actual.MoreInfoURL, msg+", MoreInfoURL is not correct")
}
