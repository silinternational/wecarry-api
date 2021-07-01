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
	if err := request.FindByUUID(tx, id.String()); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorGetRequest, api.CategoryInternal))
	}

	if !cUser.CanViewRequest(tx, request) {
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
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.Request{}, err
	}
	output.ID = request.UUID

	tx := models.Tx(ctx)
	user := models.CurrentUser(ctx)

	createdBy, err := loadRequestCreatedBy(ctx, request)
	if err != nil {
		return api.Request{}, err
	}
	output.CreatedBy = createdBy

	origin, err := loadRequestOrigin(ctx, request)
	if err != nil {
		return api.Request{}, err
	}
	output.Origin = origin

	destination, err := loadRequestDestination(ctx, request)
	if err != nil {
		return api.Request{}, err
	}
	output.Destination = destination

	provider, err := loadProvider(ctx, request)
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

	organization, err := loadRequestOrganization(ctx, request)
	if err != nil {
		return api.Request{}, err
	}
	output.Organization = organization

	meeting, err := loadRequestMeeting(ctx, request)
	if err != nil {
		return api.Request{}, err
	}
	output.Meeting = meeting

	isEditable, err := request.IsEditable(tx, user)
	if err != nil {
		return api.Request{}, err
	}
	output.IsEditable = isEditable

	return output, nil
}

// convertRequestAbridged converts model.Request into api.RequestAbridged
func convertRequestAbridged(ctx context.Context, request models.Request) (api.RequestAbridged, error) {
	var output api.RequestAbridged
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.RequestAbridged{}, err
	}
	output.ID = request.UUID

	// Hydrate nested request fields
	createdBy, err := loadRequestCreatedBy(ctx, request)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.CreatedBy = &createdBy

	origin, err := loadRequestOrigin(ctx, request)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Origin = origin

	destination, err := loadRequestDestination(ctx, request)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Destination = &destination

	provider, err := loadProvider(ctx, request)
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

func loadRequestCreatedBy(ctx context.Context, request models.Request) (api.User, error) {
	createdBy, _ := request.GetCreator(models.Tx(ctx))
	outputCreatedBy, err := convertUser(ctx, *createdBy)
	if err != nil {
		err = errors.New("error converting request created_by user: " + err.Error())
		return api.User{}, err
	}
	return outputCreatedBy, nil
}

func loadRequestOrigin(ctx context.Context, request models.Request) (*api.Location, error) {
	origin, err := request.GetOrigin(models.Tx(ctx))
	if err != nil {
		err = errors.New("error converting request origin: " + err.Error())
		return nil, err
	}

	if origin == nil {
		return nil, nil
	}

	var outputOrigin api.Location
	if err := api.ConvertToOtherType(origin, &outputOrigin); err != nil {
		err = errors.New("error converting origin to api.Location: " + err.Error())
		return nil, err
	}
	return &outputOrigin, nil
}

func loadRequestDestination(ctx context.Context, request models.Request) (api.Location, error) {
	destination, err := request.GetDestination(models.Tx(ctx))
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

func loadProvider(ctx context.Context, request models.Request) (*api.User, error) {
	tx := models.Tx(ctx)

	provider, err := request.GetProvider(tx)
	if err != nil {
		err = errors.New("error converting request provider: " + err.Error())
		return nil, err
	}

	if provider == nil {
		return nil, nil
	}

	outputProvider, err := convertUser(ctx, *provider)
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

func loadRequestOrganization(ctx context.Context, request models.Request) (api.Organization, error) {
	organization, err := request.GetOrganization(models.Tx(ctx))
	if err != nil {
		err = errors.New("error converting request organization: " + err.Error())
		return api.Organization{}, err
	}

	var outputOrganization api.Organization
	if err := api.ConvertToOtherType(organization, &outputOrganization); err != nil {
		err = errors.New("error converting organization to api.File: " + err.Error())
		return api.Organization{}, err
	}
	outputOrganization.ID = organization.UUID
	return outputOrganization, nil
}

func loadRequestMeeting(ctx context.Context, request models.Request) (*api.Meeting, error) {
	meeting, err := request.Meeting(models.Tx(ctx))
	if err != nil {
		err = errors.New("error converting request meeting: " + err.Error())
		return nil, err
	}

	if meeting == nil {
		return nil, nil
	}

	var outputMeeting api.Meeting
	if err := api.ConvertToOtherType(meeting, &outputMeeting); err != nil {
		err = errors.New("error converting meeting to api.Meeting: " + err.Error())
		return nil, err
	}
	outputMeeting.ID = meeting.UUID
	return &outputMeeting, nil
}
