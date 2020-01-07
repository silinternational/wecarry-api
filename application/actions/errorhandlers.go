package actions

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/domain"
)

type errorResponse struct {
	Error        string              `json:"error" xml:"error"`
	Code         int                 `json:"code" xml:"code,attr"`
	ContextApp   interface{}         `json:"context_app" xml:"context_app"`
	CurrentRoute interface{}         `json:"current_route" xml:"current_route"`
	Headers      http.Header         `json:"headers" xml:"headers"`
	Params       buffalo.ParamValues `json:"params" xml:"params"`
	PostedForm   url.Values          `json:"posted_form" xml:"posted_form"`
	Routes       interface{}         `json:"routes" xml:"routes"`

	// Uncomment this locally if you need to see more info
	//Context buffalo.Context `json:"context" xml:"context"`
}

func isEnvProd(env interface{}) bool {
	if env == nil {
		return false
	}

	return env.(string) == "production" || env.(string) == "prod"
}

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
	env := c.Value("env")
	c.Logger().Error(origErr)
	c.Response().WriteHeader(status)
	c.Response().Header().Set("content-type", "application/json")

	if isEnvProd(env) {
		appError := domain.AppError{
			Code: status,
			Key:  getErrorCodeFromStatus(status),
		}
		err := json.NewEncoder(c.Response()).Encode(&appError)
		return err
	}

	err := json.NewEncoder(c.Response()).Encode(&errorResponse{
		Error:        origErr.Error(),
		Code:         status,
		ContextApp:   c.Value("app"),
		CurrentRoute: c.Value("current_route"),
		Headers:      c.Request().Header,
		Params:       c.Params(),
		PostedForm:   c.Request().Form,
		Routes:       c.Value("routes"),
		//Context:      c, // Only use this in local development
	})
	return err
}
