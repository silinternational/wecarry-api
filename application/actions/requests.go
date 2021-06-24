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

// swagger:operation GET /requests Requests
//
// gets the list of requests for the current user
//
// ---
// responses:
//   '200':
//     description: requests list for the current user
//     schema:
//       "$ref": "#/definitions/Requests"
func requestsList(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	// RequestFilterParams is currently empty because the UI is not using it
	filter := models.RequestFilterParams{}

	requests := models.Requests{}
	if err := requests.FindByUser(tx, cUser, filter); err != nil {
		return domain.ReportError(c, err, "GetRequests")
	}

	// add all other fields needed --> use functions in models package to hydrate requests
	//if err := tx.Load(&requests, "CreatedBy"); err != nil {
	//	return domain.ReportError(c, err, "GetRequests")
	//}

	output,  err := convertRequestsToAPIType(c, requests)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(output))
}

func convertRequestsToAPIType(c buffalo.Context, requests []models.Request) ([]api.Request, error) {
	output := make([]api.Request, len(requests))

	for i, request := range requests {
		var err error
		output[i], err = convertRequestToAPIType(c, request)
		if err != nil {
			return []api.Request{}, err
		}
	}

	return output, nil
}

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

	// Hydrate the request's destination and origin
	tx := models.Tx(c)
	destination, err := request.GetDestination(tx)
	if err != nil {
		err = errors.New("error converting request created_by user: " + err.Error())
		return api.Request{}, err
	}

	var dest api.Location
	if err := api.ConvertToOtherType(destination, &dest); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.Request{}, err
	}
	output.Destination = &dest

	return output, nil
}
