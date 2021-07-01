package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
)

// SocialAuthConfig holds the Key and Secret for a social auth provider
type SocialAuthConfig struct{ Key, Secret string }

// Don't Modify outside of this file.
var socialAuthConfigs = map[string]SocialAuthConfig{}

// Don't Modify outside of this file.
var socialAuthOptions = []authOption{}

func init() {
	socialAuthConfigs = getSocialAuthConfigs()
	socialAuthOptions = getSocialAuthOptions(socialAuthConfigs)
}

// StrictBind hydrates a struct with values from a POST
// REMEMBER the request body must have *exported* fields.
//  Otherwise, this will give an empty result without an error.
func StrictBind(c buffalo.Context, dest interface{}) error {
	dec := json.NewDecoder(c.Request().Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dest)
}

// GetFunctionName provides the filename, line number, and function name of the caller, skipping the top `skip`
// functions on the stack.
func GetFunctionName(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "?"
	}

	fn := runtime.FuncForPC(pc)
	return fmt.Sprintf("%s:%d %s", file, line, fn.Name())
}

func httpStatusForErrCategory(cat api.ErrorCategory) int {
	switch cat {
	case api.CategoryInternal, api.CategoryDatabase:
		return http.StatusInternalServerError
	case api.CategoryForbidden, api.CategoryNotFound:
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func appErrorFromErr(err error) *api.AppError {
	aerr, ok := err.(*api.AppError)
	if ok {
		return &api.AppError{
			HttpStatus: httpStatusForErrCategory(aerr.Category),
			Key:        aerr.Key,
			Err:        aerr,
		}
	}

	return &api.AppError{
		HttpStatus: http.StatusInternalServerError,
		Key:        api.UnknownError,
		Err:        err,
	}
}

// reportError logs an error with details and renders the error with buffalo.Render.
// If the HTTP status code provided is in the 300 family, buffalo.Redirect is used instead.
func reportError(c buffalo.Context, err error) error {
	appErr, ok := err.(*api.AppError)
	if !ok {
		appErr = appErrorFromErr(err)
	}

	if appErr.Extras == nil {
		appErr.Extras = map[string]interface{}{}
	}

	appErr.Extras = api.MergeExtras([]map[string]interface{}{getExtras(c), appErr.Extras})
	appErr.Extras["function"] = GetFunctionName(2)
	appErr.Extras["key"] = appErr.Key
	appErr.Extras["status"] = appErr.HttpStatus
	appErr.Extras["redirectURL"] = appErr.RedirectURL
	appErr.Extras["method"] = c.Request().Method
	appErr.Extras["URI"] = c.Request().RequestURI
	appErr.Extras["IP"] = c.Request().RemoteAddr
	domain.Error(c, appErr.Error())

	appErr.LoadTranslatedMessage(c)

	// clear out debugging info if not in development or test
	if domain.Env.GoEnv != "development" && domain.Env.GoEnv != "test" {
		appErr.Extras = map[string]interface{}{}
	}

	if appErr.HttpStatus >= 300 && appErr.HttpStatus < 399 {
		if appErr.RedirectURL == "" {
			appErr.RedirectURL = domain.Env.UIURL + "/login?appError=" + appErr.Message
		}
		return c.Redirect(appErr.HttpStatus, appErr.RedirectURL)
	}
	return c.Render(appErr.HttpStatus, r.JSON(appErr))
}

func newExtra(c buffalo.Context, key string, e interface{}) {
	extras := getExtras(c)
	extras[key] = e
	c.Set(domain.ContextKeyExtras, extras)
}

func getExtras(c buffalo.Context) map[string]interface{} {
	extras, _ := c.Value(domain.ContextKeyExtras).(map[string]interface{})
	if extras == nil {
		extras = map[string]interface{}{}
	}
	return extras
}

func getUUIDFromParam(c buffalo.Context, param string) (uuid.UUID, error) {
	s := c.Param(param)
	id := uuid.FromStringOrNil(s)
	if id == uuid.Nil {
		newExtra(c, param, s)
		return uuid.UUID{}, &api.AppError{
			HttpStatus: http.StatusBadRequest,
			Key:        api.MustBeAValidUUID,
			Err:        fmt.Errorf("invalid %s provided: '%s'", param, s),
		}
	}
	return id, nil
}
