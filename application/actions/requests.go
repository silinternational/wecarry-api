package actions

import (
	"context"
	"errors"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/silinternational/wecarry-api/api"
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
		return reportError(c, api.NewAppError(err, api.ErrorGetRequests, api.CategoryInternal))
	}

	output, err := convertRequestsAbridged(c, requests)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// swagger:operation GET /requests/{request_id} RequestsGet
//
// gets a single request
//
// ---
// responses:
//   '200':
//     description: get a request
//     schema:
//       "$ref": "#/definitions/Request"
func requestsGet(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	id, err := getUUIDFromParam(c, "request_id")
	if err != nil {
		return reportError(c, err)
	}
	domain.NewExtra(c, "requestID", id)

	request := models.Request{}
	if err = request.FindByUUID(tx, id.String()); err != nil {
		appError := api.NewAppError(err, api.ErrorGetRequest, api.CategoryNotFound)
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return reportError(c, appError)
	}

	if !cUser.CanViewRequest(tx, request) {
		err = errors.New("user not allowed to view request")
		return reportError(c, api.NewAppError(err, api.ErrorGetRequestUserNotAllowed, api.CategoryForbidden))
	}

	output, err := convertRequest(c, request)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// converts list of model.Request into api.RequestAbridged
func convertRequestsAbridged(ctx context.Context, requests []models.Request) ([]api.RequestAbridged, error) {
	output := make([]api.RequestAbridged, len(requests))

	for i, request := range requests {
		var err error
		output[i], err = convertRequestAbridged(ctx, request)
		if err != nil {
			return []api.RequestAbridged{}, err
		}
	}

	return output, nil
}

// convertRequest converts model.Request into api.Request
func convertRequest(ctx context.Context, request models.Request) (api.Request, error) {
	var output api.Request

	if err := request.Load(ctx); err != nil {
		return output, err
	}

	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.Request{}, err
	}
	output.ID = request.UUID

	tx := models.Tx(ctx)
	user := models.CurrentUser(ctx)

	createdBy, err := convertUser(ctx, request.CreatedBy)
	if err != nil {
		return api.Request{}, err
	}
	output.CreatedBy = createdBy

	output.Destination = convertLocation(request.Destination)

	output.Origin = convertRequestOrigin(request)

	provider, err := convertProvider(ctx, request)
	if err != nil {
		return api.Request{}, err
	}
	output.Provider = provider

	photo, err := loadRequestPhoto(ctx, request)
	if err != nil {
		return api.Request{}, err
	}
	output.Photo = photo

	potentialProviders, err := loadPotentialProviders(ctx, request, user)
	if err != nil {
		return api.Request{}, err
	}
	output.PotentialProviders = potentialProviders

	output.Organization = convertOrganization(request.Organization)

	output.Meeting = convertRequestMeeting(request)

	isEditable, err := request.IsEditable(tx, user)
	if err != nil {
		return api.Request{}, err
	}
	output.IsEditable = isEditable

	return output, nil
}

// convertRequestAbridged converts model.Request into api.RequestAbridged
func convertRequestAbridged(ctx context.Context, request models.Request) (api.RequestAbridged, error) {
	if err := request.Load(ctx); err != nil {
		return api.RequestAbridged{}, err
	}

	var output api.RequestAbridged
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.RequestAbridged{}, err
	}
	output.ID = request.UUID

	// Hydrate nested request fields
	createdBy, err := convertUser(ctx, request.CreatedBy)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.CreatedBy = &createdBy

	output.Destination = convertLocation(request.Destination)

	output.Origin = convertRequestOrigin(request)

	provider, err := convertProvider(ctx, request)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Provider = provider

	photo, err := loadRequestPhoto(ctx, request)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Photo = photo

	return output, nil
}

func convertRequestOrigin(request models.Request) *api.Location {
	if !request.OriginID.Valid {
		return nil
	}

	outputOrigin := convertLocation(request.Origin)
	return &outputOrigin
}

func convertProvider(ctx context.Context, request models.Request) (*api.User, error) {
	if !request.ProviderID.Valid {
		return nil, nil
	}

	outputProvider, err := convertUser(ctx, request.Provider)
	if err != nil {
		return nil, err
	}

	return &outputProvider, nil
}

func loadPotentialProviders(ctx context.Context, request models.Request, user models.User) (api.Users, error) {
	tx := models.Tx(ctx)

	potentialProviders, err := request.GetPotentialProviders(tx, user)
	if err != nil {
		err = errors.New("error converting request potential providers: " + err.Error())
		return nil, err
	}

	outputProviders, err := convertUsers(ctx, potentialProviders)
	if err != nil {
		return nil, err
	}

	return outputProviders, nil
}

func loadRequestPhoto(ctx context.Context, request models.Request) (*api.File, error) {
	photo, err := request.GetPhoto(models.Tx(ctx))
	if err != nil {
		err = errors.New("error converting request photo: " + err.Error())
		return nil, err
	}

	if photo == nil {
		return nil, nil
	}

	var outputPhoto api.File
	if err := api.ConvertToOtherType(photo, &outputPhoto); err != nil {
		err = errors.New("error converting photo to api.File: " + err.Error())
		return nil, err
	}
	outputPhoto.ID = photo.UUID
	return &outputPhoto, nil
}

func convertRequestMeeting(request models.Request) *api.Meeting {
	if !request.MeetingID.Valid {
		return nil
	}
	meeting := convertMeeting(request.Meeting)
	return &meeting
}
