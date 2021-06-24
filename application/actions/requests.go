package actions

import (
	"context"
	"errors"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"	
)

func convertRequestToAPIType(c context.Context, request models.Request) (api.Request, error) {
	var output api.Request
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.Request{}, err
	}

	// Hydrate the request's CreatedBy user
	outputUser, err := convertUserToAPIType(c, request.CreatedBy)
	if err != nil {
		err = errors.New("error converting request created_by user: " + err.Error())
		return api.Request{}, err
	}
	output.CreatedBy = &outputUser
	output.ID = request.UUID

	return output, nil
}

func convertRequestsToAPIType(requests []models.Request) ([]api.Request, error) {
	output := make([]api.Request, len(requests))
	
	for i, request := range requests {
		var err error
		output[i], err = convertRequestToAPIType(request)
		if err != nil {
			return []api.Request{}, err
		}
	}

	return output, nil
}

// getRequests responds to Get requests at /requests

// go back to GraphQL code
func getRequests(c buffalo.Context) error {
	requests := models.Requests{}
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	filter := models.RequestFilterParams{}

	err := requests.FindByUser(tx, cUser, filter)

	// below prints an empty string
	// domain.Logger.Printf(requests[1].CreatedBy.Nickname)

	if err != nil {
		return domain.ReportError(c, err, "GetRequests")
	}

	output,  err := convertRequestsToAPIType(requests)

	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(output))
	
	// Returns all request information, but still lacking depth ("hydration") of CreatedBy,
	// 	 Organization, Provider; i.e., of nested fields
	// return c.Render(200, render.JSON(requests))
}
