package actions

import (
	"errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func convertRequestToAPIType(request models.Request) (api.Request, error) {
	var output api.Request
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to apitype.request: " + err.Error())
		return api.Request{}, err
	}

	// Hydrate the request's CreatedBy user
	var createdBy api.User
	if err := api.ConvertToOtherType(request.CreatedBy, &createdBy); err != nil {
		err = errors.New("error converting request created_by user: " + err.Error())
		return api.Request{}, err
	}
	output.CreatedBy = &createdBy

	return output, nil
}
