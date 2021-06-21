package actions

import (
	"encoding/json"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/silinternational/wecarry-api/api"
)

var httpErrorCodes = map[int]api.ErrorKey{
	http.StatusBadRequest:          api.ErrorBadRequest,
	http.StatusUnauthorized:        api.ErrorNotAuthenticated,
	http.StatusNotFound:            api.ErrorRouteNotFound,
	http.StatusMethodNotAllowed:    api.ErrorMethodNotAllowed,
	http.StatusUnprocessableEntity: api.ErrorUnprocessableEntity,
	http.StatusInternalServerError: api.ErrorInternalServerError,
}

func registerCustomErrorHandler(app *buffalo.App) {
	for i := 400; i < 600; i++ {
		app.ErrorHandlers[i] = customErrorHandler
	}
}

func getErrorCodeFromStatus(status int) api.ErrorKey {
	if s, ok := httpErrorCodes[status]; ok {
		return s
	}
	return api.ErrorUnexpectedHTTPStatus
}

func customErrorHandler(status int, origErr error, c buffalo.Context) error {
	c.Logger().Error(origErr)
	c.Response().WriteHeader(status)
	c.Response().Header().Set("content-type", "application/json")

	appError := api.AppError{
		Code: status,
		Key:  getErrorCodeFromStatus(status),
	}
	err := json.NewEncoder(c.Response()).Encode(&appError)
	return err
}
