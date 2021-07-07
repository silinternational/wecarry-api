package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func convertOrganizations(organizations models.Organizations) []api.Organization {
	output := make([]api.Organization, len(organizations))
	for i := range output {
		output[i] = convertOrganization(organizations[i])
	}

	return output
}

func convertOrganization(organization models.Organization) api.Organization {
	return api.Organization{
		ID:   organization.UUID,
		Name: organization.Name,
	}
}
