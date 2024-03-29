package actions

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
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

	output, err := models.ConvertRequest(c, request)
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
		return reportError(c, err)
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
		CreatedByID: models.CurrentUser(ctx).ID,
		Description: input.Description,
		Kilograms:   input.Kilograms,
		Size:        models.RequestSize(input.Size),
		Status:      models.RequestStatusOpen,
		Title:       input.Title,
		Visibility:  models.RequestVisibility(input.Visibility),
	}

	if input.NeededBefore.Valid {
		if err := addNeededBeforeToRequest(tx, input.NeededBefore, &request); err != nil {
			return request, err
		}
	}

	if input.MeetingID.Valid {
		if err := addMeetingIDToRequest(tx, input.MeetingID, &request); err != nil {
			return request, err
		}
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
		if err := attachPhotoToRequest(tx, input.PhotoID, &request); err != nil {
			return request, err
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
		return reportError(c, err)
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
		if err := attachPhotoToRequest(tx, input.PhotoID, &request); err != nil {
			return request, err
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

	if input.MeetingID.Valid {
		if err := addMeetingIDToRequest(tx, input.MeetingID, &request); err != nil {
			return request, err
		}
	} else {
		request.MeetingID = nulls.Int{}
	}

	if input.NeededBefore.Valid {
		if err := addNeededBeforeToRequest(tx, input.NeededBefore, &request); err != nil {
			return request, err
		}
	} else {
		request.NeededBefore = nulls.Time{}
	}

	return request, nil
}

func addNeededBeforeToRequest(tx *pop.Connection, neededBefore nulls.String, request *models.Request) error {
	n, err := time.Parse(domain.DateFormat, neededBefore.String)
	if err != nil {
		err = errors.New("failed to parse NeededBefore, " + err.Error())
		appErr := api.NewAppError(err, api.ErrorCreateRequestInvalidDate, api.CategoryUser)
		return appErr
	}
	request.NeededBefore = nulls.NewTime(n)
	return nil
}

func addMeetingIDToRequest(tx *pop.Connection, meetingID nulls.UUID, request *models.Request) error {
	var meeting models.Meeting
	if err := meeting.FindByUUID(tx, meetingID.UUID.String()); err != nil {
		err = errors.New("meeting ID not found, " + err.Error())
		appErr := api.NewAppError(err, api.ErrorRequestMeetingIDNotFound, api.CategoryUser)
		if domain.IsOtherThanNoRows(err) {
			appErr.Category = api.CategoryDatabase
			return appErr
		}
		return appErr
	}

	request.MeetingID = nulls.NewInt(meeting.ID)
	return nil
}

func attachPhotoToRequest(tx *pop.Connection, photoID nulls.UUID, request *models.Request) error {
	if _, err := request.AttachPhoto(tx, photoID.UUID.String()); err != nil {
		err = errors.New("request photo file ID not found, " + err.Error())
		appErr := api.NewAppError(err, api.ErrorRequestPhotoIDNotFound, api.CategoryUser)
		if domain.IsOtherThanNoRows(err) {
			appErr.Category = api.CategoryDatabase
		}
		return appErr
	}

	return nil
}

// swagger:operation POST /requests/{request_id}/potentialprovider Requests AddMeAsPotentialProvider
//
// Adds the current user as a potential provider to the request
//
// ---
// responses:
//   '200':
//     description: a request
//     schema:
//       "$ref": "#/definitions/Request"
func requestsAddMeAsPotentialProvider(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	id, err := getUUIDFromParam(c, "request_id")
	if err != nil {
		return reportError(c, err)
	}
	domain.NewExtra(c, "requestID", id)

	request := models.Request{}
	if err = request.AddUserAsPotentialProvider(tx, id.String(), cUser); err != nil {
		appError := api.NewAppError(err, api.ErrorGetRequest, api.CategoryInternal)
		if strings.Contains(err.Error(), "error creating potential provider: unique_together") {
			appError.Key = api.ErrorAddPotentialProviderDuplicate
			appError.Category = api.CategoryUser
		} else if !domain.IsOtherThanNoRows(err) || strings.Contains(err.Error(), "may not view request") {
			appError.Category = api.CategoryNotFound
		}
		return reportError(c, appError)
	}

	output, err := models.ConvertRequest(c, request)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// swagger:operation DELETE /requests/{request_id}/potentialprovider/{user_id} Requests RejectPotentialProvider
//
// Requester removes a potential provider attached to their request
//
// ---
// responses:
//   '204':
//     description: OK but no content in response
func requestsRejectPotentialProvider(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	requestID, err := getUUIDFromParam(c, "request_id")
	if err != nil {
		return reportError(c, err)
	}
	domain.NewExtra(c, "requestID", requestID)

	userID, err := getUUIDFromParam(c, "user_id")
	if err != nil {
		return reportError(c, err)
	}
	domain.NewExtra(c, "requestID", userID)

	var provider models.PotentialProvider
	if err := provider.FindWithRequestUUIDAndUserUUID(tx, requestID.String(), userID.String(), cUser); err != nil {
		appError := api.NewAppError(err, api.ErrorRejectPotentialProviderForbidden, api.CategoryForbidden)
		if strings.Contains(err.Error(), "unable to find User") {
			appError.Key = api.ErrorRejectPotentialProviderFindUser
		} else if strings.Contains(err.Error(), "unable to find Request") {
			appError.Key = api.ErrorGetRequest
			appError.Category = api.CategoryUser
		} else if strings.Contains(err.Error(), "unable to find PotentialProvider") {
			appError.Key = api.ErrorRejectPotentialProviderFindProvider
			appError.Category = api.CategoryUser
		}

		return reportError(c, appError)
	}

	if err := provider.Destroy(tx); err != nil {
		appError := api.NewAppError(err, api.ErrorRejectPotentialProviderDestroyIt, api.CategoryInternal)
		return reportError(c, appError)
	}

	return c.Render(http.StatusNoContent, nil)
}

// swagger:operation DELETE /requests/{request_id}/potentialprovider Requests RemoveMeAsPotentialProvider
//
// Removes the current user as a potential provider to the request
//
// ---
// responses:
//   '204':
//     description: OK but no content in response
func requestsRemoveMeAsPotentialProvider(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	id, err := getUUIDFromParam(c, "request_id")
	if err != nil {
		return reportError(c, err)
	}
	domain.NewExtra(c, "requestID", id)

	var provider models.PotentialProvider
	if err := provider.FindWithRequestUUIDAndUserUUID(tx, id.String(), cUser.UUID.String(), cUser); err != nil {
		appError := api.NewAppError(err, api.ErrorRemoveMeAsPotentialProviderForbidden, api.CategoryForbidden)
		if strings.Contains(err.Error(), "unable to find User") {
			appError.Key = api.ErrorRemoveMeAsPotentialProviderFindUser
		} else if strings.Contains(err.Error(), "unable to find Request") {
			appError.Key = api.ErrorGetRequest
			appError.Category = api.CategoryUser
		} else if strings.Contains(err.Error(), "unable to find PotentialProvider") {
			appError.Key = api.ErrorRemoveMeAsPotentialProviderFindProvider
			appError.Category = api.CategoryUser
		}

		return reportError(c, appError)
	}

	if err := provider.Destroy(tx); err != nil {
		appError := api.NewAppError(err, api.ErrorRemoveMeAsPotentialProviderDestroyIt, api.CategoryInternal)
		return reportError(c, appError)
	}

	return c.Render(http.StatusNoContent, nil)
}

// swagger:operation PUT /requests/{request_id}/status Requests RequestsUpdateStatus
//
// update the status of a request
//
// ---
// parameters:
//   - name: RequestUpdateStatusInput
//     in: body
//     required: true
//     description: input object
//     schema:
//       "$ref": "#/definitions/RequestUpdateStatusInput"
//
// responses:
//   '200':
//     description: the request
//     schema:
//       "$ref": "#/definitions/Request"
func requestsUpdateStatus(c buffalo.Context) error {
	var input api.RequestUpdateStatusInput
	if err := StrictBind(c, &input); err != nil {
		return reportError(c, err)
	}

	requestID, err := getUUIDFromParam(c, "request_id")
	if err != nil {
		return reportError(c, err)
	}

	tx := models.Tx(c)
	var request models.Request
	if err := request.FindByUUID(tx, requestID.String()); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorUpdateRequestStatusNotFound, api.CategoryNotFound))
	}

	cUser := models.CurrentUser(c)
	domain.NewExtra(c, "oldStatus", request.Status)
	domain.NewExtra(c, "newStatus", input.Status)

	if !cUser.CanUpdateRequestStatus(request, models.RequestStatus(input.Status)) {
		err = errors.New("not allowed to change request status")
		return reportError(c, api.NewAppError(err, api.ErrorUpdateRequestStatusBadStatus, api.CategoryUser))
	}

	if err = request.SetProviderWithStatus(tx, models.RequestStatus(input.Status), input.ProviderUserID); err != nil {
		err = errors.New("error setting provider with status: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorUpdateRequestStatusBadProvider, api.CategoryUser))
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
