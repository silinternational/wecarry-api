package actions

import (
	"context"
	"errors"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v5"

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

	output,  err := convertRequestsToAPIType(c, requests)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(output))
}

// converts list of model.Request into api.RequestAbridged
func convertRequestsToAPIType(ctx context.Context, requests []models.Request) ([]api.RequestAbridged, error) {
	output := make([]api.RequestAbridged, len(requests))

	for i, request := range requests {
		var err error
		output[i], err = convertRequestToAPIType(ctx, request)
		if err != nil {
			return []api.RequestAbridged{}, err
		}
	}

	return output, nil
}

// converts model.Request into api.RequestAbridged
func convertRequestToAPIType(ctx context.Context, request models.Request) (api.RequestAbridged, error) {
	var output api.RequestAbridged
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.RequestAbridged{}, err
	}
	output.ID = request.UUID

	// Hydrate nested request fields
	tx := models.Tx(ctx)
	createdBy, err := hydrateCreatedBy(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.CreatedBy = &createdBy
	origin, err := hydrateOrigin(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Origin = &origin
	destination, err := hydrateDestination(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Destination = &destination
	provider, err := hydrateProvider(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Provider = &provider
	photo, err := hydratePhoto(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Photo = &photo
	return output, nil
}

func hydrateCreatedBy (ctx context.Context, request models.Request, tx *pop.Connection) (api.User, error) {
	createdBy, _ := request.GetCreator(tx)
	outputCreatedBy, err := convertUserToAPIType(ctx, *createdBy)
	if err != nil {
		err = errors.New("error converting request created_by user: " + err.Error())
		return api.User{}, err
	}
	return outputCreatedBy, nil
}

func hydrateOrigin (ctx context.Context, request models.Request, tx *pop.Connection) (api.Location, error) {
	origin, err := request.GetOrigin(tx)
	if err != nil {
		err = errors.New("error converting request origin: " + err.Error())
		return api.Location{}, err
	}
	var outputOrigin api.Location
	if err := api.ConvertToOtherType(origin, &outputOrigin); err != nil {
		err = errors.New("error converting origin to api.Location: " + err.Error())
		return api.Location{}, err
	}
	return outputOrigin, nil
}

func hydrateDestination (ctx context.Context, request models.Request, tx *pop.Connection) (api.Location, error) {
	destination, err := request.GetDestination(tx)
	if err != nil {
		err = errors.New("error converting request destination: " + err.Error())
		return api.Location{}, err
	}
	var outputDestination api.Location
	if err := api.ConvertToOtherType(destination, &outputDestination); err != nil {
		err = errors.New("error converting destination to api.Location: " + err.Error())
		return api.Location{}, err
	}
	return outputDestination, nil
}

func hydrateProvider (ctx context.Context, request models.Request, tx *pop.Connection) (api.User, error) {
	provider, err := request.GetProvider(tx)
	if err != nil {
		err = errors.New("error converting request provider: " + err.Error())
		return api.User{}, err
	}
	var outputProvider api.User
	if err := api.ConvertToOtherType(provider, &outputProvider); err != nil {
		err = errors.New("error converting provider to api.User: " + err.Error())
		return api.User{}, err
	}
	if provider != nil {
		outputProvider.ID = provider.UUID
	}
	return outputProvider, nil
}

func hydratePhoto (ctx context.Context, request models.Request, tx *pop.Connection) (api.File, error) {
	photo, err := request.GetPhoto(tx)
	if err != nil {
		err = errors.New("error converting request photo: " + err.Error())
		return api.File{}, err
	}
	var outputPhoto api.File
	if err := api.ConvertToOtherType(photo, &outputPhoto); err != nil {
		err = errors.New("error converting photo to api.File: " + err.Error())
		return api.File{}, err
	}
	if photo != nil {
		outputPhoto.ID = photo.UUID
	}
	return outputPhoto, nil
}

