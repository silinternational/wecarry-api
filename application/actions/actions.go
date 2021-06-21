package actions

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/apitypes"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/silinternational/wecarry-api/wcerror"
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

func httpStatusForErrCategory(cat wcerror.ErrorCategory) int {
	switch cat {
	case wcerror.CategoryInternal, wcerror.CategoryDatabase:
		return http.StatusInternalServerError
	case wcerror.CategoryForbidden, wcerror.CategoryNotFound:
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func appErrorFromErr(err error) *apitypes.AppError {
	terr, ok := err.(*wcerror.WCerror)
	if ok {
		return &apitypes.AppError{
			HttpStatus: httpStatusForErrCategory(terr.Category),
			Key:        terr.Key,
			DebugMsg:   terr.Error(),
		}
	}

	return &apitypes.AppError{
		HttpStatus: http.StatusInternalServerError,
		Key:        wcerror.UnknownError,
		DebugMsg:   err.Error(),
	}
}

// reportError logs an error with details and renders the error with buffalo.Render.
// If the HTTP status code provided is in the 300 family, buffalo.Redirect is used instead.
func reportError(c buffalo.Context, err error) error {
	appErr, ok := err.(*apitypes.AppError)
	if !ok {
		appErr = appErrorFromErr(err)
	}

	if appErr.Extras == nil {
		appErr.Extras = map[string]interface{}{}
	}

	appErr.Extras = domain.MergeExtras([]map[string]interface{}{getExtras(c), appErr.Extras})
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
		appErr.DebugMsg = ""
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
