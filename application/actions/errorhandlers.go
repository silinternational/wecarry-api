package actions

import (
	"encoding/json"
	"net/http"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/domain"
)

func getErrorCodeFromStatus(status int) string {
	var codes = map[int]string{
		http.StatusBadRequest:          domain.ErrorBadRequest,
		http.StatusUnauthorized:        domain.ErrorNotAuthenticated,
		http.StatusNotFound:            domain.ErrorRouteNotFound,
		http.StatusMethodNotAllowed:    domain.ErrorMethodNotAllowed,
		http.StatusUnprocessableEntity: domain.ErrorUnprocessableEntity,
		http.StatusInternalServerError: domain.ErrorInternalServerError,
	}
	if s, ok := codes[status]; ok {
		return s
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
