package actions

import (
	"context"
	"errors"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/cache"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// swagger:operation GET /requests Requests ListRequests
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

	orgs, err := cUser.GetOrganizations(tx)
	if err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorGetRequests, api.CategoryInternal))
	}

	requestsList, err := cache.GetVisibleRequests(c, orgs)
	if err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorGetRequests, api.CategoryInternal))
	}

	return c.Render(200, render.JSON(requestsList))
}

// swagger:operation GET /requests/{request_id} Requests GetRequest
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

	output, err := models.ConvertRequestAbridged(c, request)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// swagger:operation POST /requests Requests RequestsCreate
//
// create a new request
//
// ---
// parameters:
//   - name: RequestCreateInput
//     in: body
//     required: true
//     description: input object
//     schema:
//       "$ref": "#/definitions/RequestCreateInput"
//
// responses:
//   '200':
//     description: the created request
//     schema:
//       "$ref": "#/definitions/Request"
func requestsCreate(c buffalo.Context) error {
	var input api.RequestCreateInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal data into RequestCreateInput, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	tx := models.Tx(c)

	request, err := convertRequestCreateInput(c, input)
	if err != nil {
		return reportError(c, err)
	}

	if err = request.Create(tx); err != nil {
		// TODO: discern whether err is a database error or a validation error or ???
		return reportError(c, api.NewAppError(err, api.ErrorCreateRequest, api.CategoryUser))
	}

	output, err := models.ConvertRequest(c, request)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// convertRequestCreateInput creates a new `Request` from a `RequestCreateInput`. All properties that are not `nil` are
// set to the value provided
// TODO: move the actual DB access into Request.Update()
func convertRequestCreateInput(ctx context.Context, input api.RequestCreateInput) (models.Request, error) {
	tx := models.Tx(ctx)

	request := models.Request{
		CreatedByID:  models.CurrentUser(ctx).ID,
		Description:  input.Description,
		Kilograms:    input.Kilograms,
		NeededBefore: input.NeededBefore,
		Size:         models.RequestSize(input.Size),
		Status:       models.RequestStatusOpen,
		Title:        input.Title,
		Visibility:   models.RequestVisibility(input.Visibility),
	}

	if input.OrganizationID != uuid.Nil {
		var org models.Organization
		if err := org.FindByUUID(tx, input.OrganizationID.String()); err != nil {
			err = errors.New("organization ID not found, " + err.Error())
			appErr := api.NewAppError(err, api.ErrorCreateRequestOrgIDNotFound, api.CategoryUser)
			if domain.IsOtherThanNoRows(err) {
				appErr.Category = api.CategoryDatabase
				return request, appErr
			}
			return request, appErr
		}
		request.OrganizationID = org.ID
	}

	if input.PhotoID.Valid {
		if _, err := request.AttachPhoto(tx, input.PhotoID.UUID.String()); err != nil {
			err = errors.New("file ID not found, " + err.Error())
			appErr := api.NewAppError(err, api.ErrorCreateRequestPhotoIDNotFound, api.CategoryUser)
			if domain.IsOtherThanNoRows(err) {
				appErr.Category = api.CategoryDatabase
			}
			return request, appErr
		}
	}

	destination := models.ConvertLocationInput(input.Destination)
	if err := destination.Create(tx); err != nil {
		return request, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryUser)
	}
	request.DestinationID = destination.ID

	if input.Origin != nil {
		origin := models.ConvertLocationInput(*input.Origin)
		if err := origin.Create(tx); err != nil {
			return request, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryUser)
		}
		request.OriginID = nulls.NewInt(origin.ID)
	}

	return request, nil
}

// swagger:operation PUT /requests/{request_id} Requests RequestsUpdate
//
// update a request
//
// ---
// parameters:
//   - name: RequestUpdateInput
//     in: body
//     required: true
//     description: input object
//     schema:
//       "$ref": "#/definitions/RequestUpdateInput"
//
// responses:
//   '200':
//     description: the request
//     schema:
//       "$ref": "#/definitions/Request"
func requestsUpdate(c buffalo.Context) error {
	var input api.RequestUpdateInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal data into RequestCreateInput, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	tx := models.Tx(c)

	requestID, err := getUUIDFromParam(c, "request_id")
	if err != nil {
		return reportError(c, err)
	}

	request, err := convertRequestUpdateInput(c, input, requestID.String())
	if err != nil {
		return reportError(c, err)
	}

	if err = request.Update(tx); err != nil {
		appError := api.NewAppError(err, api.ErrorUpdateRequest, api.CategoryUser)
		// TODO: check for validation errors
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return reportError(c, appError)
	}

	output, err := models.ConvertRequest(c, request)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// convertRequestUpdateInput updates a `Request` from a `RequestUpdateInput`, with the values in
// the database as default for any property not specified in the input.
// TODO: move the actual DB access into Request.Update() -- Note that this requires a change of function signature so that Update has the before and after File and Location IDs.
func convertRequestUpdateInput(ctx context.Context, input api.RequestUpdateInput, id string) (models.Request, error) {
	tx := models.Tx(ctx)

	var request models.Request
	if err := request.FindByUUID(tx, id); err != nil {
		appError := api.NewAppError(err, api.ErrorUpdateRequest, api.CategoryNotFound)
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return request, appError
	}

	if input.PhotoID.Valid {
		if _, err := request.AttachPhoto(tx, input.PhotoID.UUID.String()); err != nil {
			err = errors.New("request photo file ID not found, " + err.Error())
			appErr := api.NewAppError(err, api.ErrorUpdateRequestPhotoIDNotFound, api.CategoryUser)
			if domain.IsOtherThanNoRows(err) {
				appErr.Category = api.CategoryDatabase
			}
			return request, appErr
		}
	} else {
		if err := request.RemoveFile(tx); err != nil {
			return request, err
		}
	}

	if input.Destination != nil {
		destination := models.ConvertLocationInput(*input.Destination)
		if err := request.SetDestination(tx, destination); err != nil {
			return request, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryUser)
		}
	}

	if input.Origin != nil {
		origin := models.ConvertLocationInput(*input.Origin)
		if err := request.SetOrigin(tx, origin); err != nil {
			return request, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryUser)
		}
	} else {
		if err := request.RemoveOrigin(tx); err != nil {
			return request, api.NewAppError(err, api.ErrorLocationDeleteFailure, api.CategoryInternal)
		}
	}

	if input.Size != nil {
		request.Size = models.RequestSize(*input.Size)
	}

	if input.Title != nil {
		request.Title = *input.Title
	}

	if input.Visibility != nil {
		request.Visibility = models.RequestVisibility(*input.Visibility)
	}

	request.Description = input.Description

	request.Kilograms = input.Kilograms

	request.NeededBefore = input.NeededBefore

	return request, nil
}
