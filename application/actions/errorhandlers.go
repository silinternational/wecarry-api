package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
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

func customErrorHandler(status int, origErr error, c buffalo.Context) error {
	c.Logger().Error(origErr)
	c.Response().WriteHeader(status)
	c.Response().Header().Set("content-type", "application/json")

	if status >= 500 && domain.Env.GoEnv == "development" {
		debug.PrintStack()
	}

	appError := api.AppError{
		Code: status,
		Key:  getErrorCodeFromStatus(status),
	}
	appError.LoadTranslatedMessage(c)
	if domain.Env.GoEnv == "development" {
		appError.DebugMsg = fmt.Sprintf("(%T) %s", origErr, origErr)
	}

	address, _ := getClientIPAddress(c)
	e := log.WithFields(map[string]any{
		domain.ExtrasKey:    appError.Key,
		domain.ExtrasStatus: status,
		domain.ExtrasMethod: c.Request().Method,
		"uri":               c.Request().RequestURI,
		"ip":                address,
	})
	if status >= 500 {
		e.Errorf(origErr.Error())
	} else {
		e.Warningf(origErr.Error())
	}

	err := json.NewEncoder(c.Response()).Encode(&appError)
	return err
}

func getErrorCodeFromStatus(status int) api.ErrorKey {
	if s, ok := httpErrorCodes[status]; ok {
		return s
	}
	return api.ErrorGenericInternalServer
}
