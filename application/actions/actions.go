package actions

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
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
//
//	Otherwise, this will give an empty result without an error.
func StrictBind(c buffalo.Context, dest interface{}) error {
	dec := json.NewDecoder(c.Request().Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser)
	}
	return nil
}

// appErrorFromErr is used by reportError to convert a generic error to an AppError
func appErrorFromErr(err error) *api.AppError {
	appErr, ok := err.(*api.AppError)
	if ok {
		return appErr
	}

	return &api.AppError{
		HttpStatus: http.StatusInternalServerError,
		Key:        api.ErrorUnknown,
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
	appErr.SetHttpStatusFromCategory()

	if appErr.Extras == nil {
		appErr.Extras = map[string]interface{}{}
	}

	appErr.Extras = api.MergeExtras([]map[string]interface{}{getExtras(c), appErr.Extras})
	appErr.Extras["function"] = domain.GetFunctionName(2)
	appErr.Extras[domain.ExtrasKey] = appErr.Key
	appErr.Extras[domain.ExtrasStatus] = appErr.HttpStatus
	appErr.Extras["redirectURL"] = appErr.RedirectURL
	appErr.Extras[domain.ExtrasMethod] = c.Request().Method
	appErr.Extras[domain.ExtrasURI] = c.Request().RequestURI

	address, _ := getClientIPAddress(c)
	appErr.Extras[domain.ExtrasIP] = address

	entry := log.WithContext(c).WithFields(appErr.Extras)
	switch appErr.Category {
	case api.CategoryUser:
		entry.Warning(err)
	default:
		entry.Error(err)
	}

	appErr.LoadTranslatedMessage(c)

	// clear out debugging info if not in development or test
	if domain.Env.GoEnv == "development" || domain.Env.GoEnv == "test" {
		appErr.DebugMsg = appErr.Err.Error()
	} else {
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

// newExtra inserts a new data item into the context for use in debugging
func newExtra(c buffalo.Context, key string, e interface{}) {
	extras := getExtras(c)
	extras[key] = e
	c.Set(domain.ContextKeyExtras, extras)
}

// getExtras obtains the map of extra data for insertion into a log message
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
		err := fmt.Errorf("invalid %s provided: '%s'", param, s)
		return uuid.UUID{}, api.NewAppError(err, api.ErrorMustBeAValidUUID, api.CategoryUser)
	}
	return id, nil
}
