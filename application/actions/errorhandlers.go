package actions

import (
	"encoding/json"
	"net/http"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/domain"
)

func getErrorCodeFromStatus(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return domain.ErrorNotAuthenticated
	case http.StatusNotFound:
		return domain.ErrorRouteNotFound
	case http.StatusMethodNotAllowed:
		return domain.ErrorMethodNotAllowed
	case http.StatusInternalServerError:
		return domain.ErrorInternalServerError
	}

	return domain.ErrorUnexpectedHTTPStatus
}

func customErrorHandler(status int, origErr error, c buffalo.Context) error {
	c.Logger().Error(origErr)
	c.Response().WriteHeader(status)
	c.Response().Header().Set("content-type", "application/json")

	appError := domain.AppError{
		Code: status,
		Key:  getErrorCodeFromStatus(status),
	}
	err := json.NewEncoder(c.Response()).Encode(&appError)
	return err
}
