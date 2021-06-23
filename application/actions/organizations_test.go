package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) Test_convertOrganizationToAPIType() {
	u := domain.GetUUID()

	organization := models.Organization{
		UUID: u,
		Name: "test org",
	}
	want := api.Organization{
		ID:   u,
		Name: "test org",
	}
	got, err := convertOrganizationToAPIType(organization)
	as.NoError(err)
	as.Equal(want, got)
}
