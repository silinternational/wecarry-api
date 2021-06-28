package actions

import (
	"context"
	"errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func convertRequestToAPIType(ctx context.Context, request models.Request) (api.Request, error) {
	var output api.Request
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.Request{}, err
	}

	// Hydrate the request's CreatedBy user
	outputUser, err := convertUser(ctx, request.CreatedBy)
	if err != nil {
		err = errors.New("error converting request created_by user: " + err.Error())
		return api.Request{}, err
	}
	output.CreatedBy = &outputUser
	output.ID = request.UUID

	return output, nil
}
