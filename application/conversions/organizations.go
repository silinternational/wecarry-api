package conversions

import (
	"errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func convertOrganizationsToAPIType(organizations models.Organizations) ([]api.Organization, error) {
	output := make([]api.Organization, len(organizations))
	for i := range output {
		o, err := convertOrganizationToAPIType(organizations[i])
		if err != nil {
			return nil, err
		}
		output[i] = o
	}

	return output, nil
}

func convertOrganizationToAPIType(organization models.Organization) (api.Organization, error) {
	var output api.Organization
	if err := api.ConvertToOtherType(organization, &output); err != nil {
		err = errors.New("error converting organization to api.organization: " + err.Error())
		return api.Organization{}, err
	}
	output.ID = organization.UUID

	return output, nil
}
