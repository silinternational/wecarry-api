package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) Test_convertOrganization() {
	u := domain.GetUUID()

	organization := models.Organization{
		UUID: u,
		Name: "test org",
	}
	got := models.ConvertOrganization(organization)
	as.verifyOrganization(organization, got, "Organization is not correct")
}

func (as *ActionSuite) verifyOrganization(expected models.Organization, actual api.Organization, msg string) {
	as.Equal(expected.UUID, actual.ID, msg+", ID is not correct")
	as.Equal(expected.Name, actual.Name, msg+", Name is not correct")
}
