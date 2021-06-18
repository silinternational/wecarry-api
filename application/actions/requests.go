package actions

import (
	"errors"

	"github.com/silinternational/wecarry-api/apitypes"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func convertRequestToAPIType(request models.Request) (apitypes.Request, error) {
	var output apitypes.Request
	if err := domain.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to apitype.request: " + err.Error())
		return apitypes.Request{}, err
	}

	// Hydrate the request's CreatedBy user
	var createdBy apitypes.User
	if err := domain.ConvertToOtherType(request.CreatedBy, &createdBy); err != nil {
		err = errors.New("error converting request created_by user: " + err.Error())
		return apitypes.Request{}, err
	}
	output.CreatedBy = &createdBy

	return output, nil
}
