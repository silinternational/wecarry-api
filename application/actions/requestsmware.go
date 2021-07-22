package actions

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gobuffalo/pop/v5"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"

	"github.com/gobuffalo/buffalo"
)

// getRequestAuthz is used to get the desired Request and
// check authorization
func getRequestAuthz(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		rID := c.Param(requestIDKey)
		if rID == "" {
			return next(c)
		}

		cUser := models.CurrentUser(c)
		domain.NewExtra(c, "currentUserID", cUser.UUID)

		requestID, err := getUUIDFromParam(c, "request_id")
		if err != nil {
			return reportError(c, err)
		}

		domain.NewExtra(c, "requestID", requestID)

		tx := models.Tx(c)
		request := models.Request{}
		if err := request.FindByUUIDForCurrentUser(tx, requestID.String(), cUser); err != nil {
			appError := api.NewAppError(err, api.ErrorGetRequest, api.CategoryNotFound)
			if domain.IsOtherThanNoRows(err) && !strings.Contains(err.Error(), "unauthorized") {
				appError.Category = api.CategoryInternal
			}
			return reportError(c, appError)
		}

		reqPath := c.Request().URL.Path
		isStatusChange := strings.Contains(reqPath, "status")

		// Ensure user is allowed to update the request
		if c.Request().Method == http.MethodPut && !isStatusChange {
			if err := dealWithRequestPut(c, tx, request, cUser); err != nil {
				return reportError(c, err)
			}
		}

		data := handlerData{
			User:     cUser,
			ModelObj: request,
		}

		setHandlerData(c, data)

		return next(c)
	}
}

func dealWithRequestPut(c buffalo.Context, tx *pop.Connection, request models.Request, cUser models.User) error {
	ok, err := request.IsEditable(tx, cUser)
	if err != nil {
		appError := api.NewAppError(err, api.ErrorUpdateRequest, api.CategoryInternal)
		return reportError(c, appError)
	}
	if !ok {
		err := errors.New("User is not allowed to edit request")
		appError := api.NewAppError(err, api.ErrorUpdateRequest, api.CategoryForbidden)
		return reportError(c, appError)
	}

	return nil
}

type handlerRequest struct {
	User    models.User
	Request models.Request
}

func getHandlerRequestData(c buffalo.Context) (handlerRequest, error) {
	hdata, ok := c.Value(handlerDataKey).(handlerData)
	if !ok {
		err := errors.New("handlerData not found in context")
		return handlerRequest{}, api.NewAppError(err, api.ErrorUnknown, api.CategoryInternal)
	}

	req, ok := hdata.ModelObj.(models.Request)
	if !ok {
		err := errors.New("handlerData did not contain a Request object")
		return handlerRequest{}, api.NewAppError(err, api.ErrorUnknown, api.CategoryInternal)
	}

	handlerReq := handlerRequest{
		User:    hdata.User,
		Request: req,
	}

	return handlerReq, nil
}
