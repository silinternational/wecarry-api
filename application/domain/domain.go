package domain

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/rollbar/rollbar-go"

	"github.com/gobuffalo/buffalo"

	uuid2 "github.com/gofrs/uuid"
)

const (
	ErrorLevelWarn              = "warn"
	ErrorLevelError             = "error"
	ErrorLevelCritical          = "critical"
	AdminRoleSuperDuperAdmin    = "SuperDuperAdmin"
	AdminRoleSalesAdmin         = "SalesAdmin"
	EmptyUUID                   = "00000000-0000-0000-0000-000000000000"
	DateFormat                  = "2006-01-02"
	MaxFileSize                 = 1 << 21 // 10 Mebibytes
	AccessTokenLifetimeSeconds  = 3600
	DateTimeFormat              = "2006-01-02 15:04:05"
	NewMessageNotificationDelay = 10 * time.Minute
)

// Environment Variables
const (
	UIURLEnv                      = "UI_URL"
	AccessTokenLifetimeSecondsEnv = "ACCESS_TOKEN_LIFETIME_SECONDS"
	SendGridAPIKeyEnv             = "SENDGRID_API_KEY"
	EmailServiceEnv               = "EMAIL_SERVICE"
	MobileServiceEnv              = "MOBILE_SERVICE"
)

// Event Kinds
const (
	EventApiUserCreated      = "api:user:created"
	EventApiAuthUserLoggedIn = "api:auth:user:loggedin"
	EventApiMessageCreated   = "api:message:created"
)

// Event and Job argument names
const (
	ArgMessageID = "message_id"
)

// Notification Message Template Names
const (
	MessageTemplateNewMessage = "new_message"
	MessageTemplateNewRequest = "new_request"
)

var Logger log.Logger
var ErrLogger log.Logger

func init() {
	Logger.SetOutput(os.Stdout)
	ErrLogger.SetOutput(os.Stderr)
}

type AppError struct {
	Code    string `json:"Code"`
	Message string `json:"Message,omitempty"`
}

// GetFirstStringFromSlice returns the first string in the given slice, or an empty
// string if the slice is empty or nil.
func GetFirstStringFromSlice(strSlice []string) string {
	if len(strSlice) > 0 {
		return strSlice[0]
	}

	return ""
}

// GetBearerTokenFromRequest obtains the token from an Authorization header beginning
// with "Bearer". If not found, an empty string is returned.
func GetBearerTokenFromRequest(r *http.Request) string {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return ""
	}

	re := regexp.MustCompile(`^(?i)Bearer (.*)$`)
	matches := re.FindSubmatch([]byte(authorizationHeader))
	if len(matches) < 2 {
		return ""
	}

	return string(matches[1])
}

// GetSubPartKeyValues parses a string of parameter/value pairs delimited by `outerDelimiter`.
// Each pair is split by `innerDelimiter` and returned as entries in a map[string]string.
// Example: "param1=value1-param2=value2" produces {"param1": "value1", "param2": "value2"}
// if `outerDelimiter` is "-" and `innerDelimiter` is "=".
func GetSubPartKeyValues(inString, outerDelimiter, innerDelimiter string) map[string]string {
	keyValues := map[string]string{}
	allPairs := strings.Split(inString, outerDelimiter)

	for _, p := range allPairs {
		pParts := strings.Split(p, innerDelimiter)
		if len(pParts) == 2 {
			keyValues[pParts[0]] = pParts[1]
		}
	}

	return keyValues
}

// ConvertTimeToStringPtr is intended to convert the
// CreatedAt and UpdatedAt fields of database objects
// to pointers to strings to populate the same gqlgen fields
func ConvertTimeToStringPtr(inTime time.Time) *string {
	inTimeStr := inTime.Format(time.RFC3339)
	return &inTimeStr
}

// ConvertStrPtrToString dereferences a string pointer and returns
// the result. In case nil is given, an empty string is returned.
func ConvertStrPtrToString(inPtr *string) string {
	if inPtr == nil {
		return ""
	}

	return *inPtr
}

// GetCurrentTime returns a string of the current date and time
// based on the default DateTimeFormat
func GetCurrentTime() string {
	return time.Now().Format(DateTimeFormat)
}

// GetUuid creates a new, unique version 4 (random) UUID and returns it
// as a uuid2.UUID. Errors are ignored.
func GetUuid() uuid2.UUID {
	uuid, err := uuid2.NewV4()
	if err != nil {
		ErrLogger.Printf("error creating new uuid2 ... %v", err)
	}
	return uuid
}

// ConvertStringPtrToDate uses time.Parse to convert a date in yyyy-mm-dd
// format into a time.Time object. If nil is provided, the default value
// for time.Time is returned.
func ConvertStringPtrToDate(inPtr *string) (time.Time, error) {
	if inPtr == nil || *inPtr == "" {
		return time.Time{}, nil
	}

	return time.Parse(DateFormat, *inPtr)
}

// IsStringInSlice iterates over a slice of strings, looking for the given
// string. If found, true is returned. Otherwise, false is returned.
func IsStringInSlice(needle string, haystack []string) bool {
	for _, hs := range haystack {
		if needle == hs {
			return true
		}
	}

	return false
}

func EmailDomain(email string) string {
	// If email includes @ it is full email address, otherwise it is just domain
	if strings.Contains(email, "@") {
		parts := strings.Split(email, "@")
		return parts[1]
	} else {
		return email
	}
}

func RollbarMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		client := rollbar.New(
			envy.Get("ROLLBAR_TOKEN", ""),
			envy.Get("GO_ENV", "development"),
			"",
			"",
			envy.Get("ROLLBAR_SERVER_ROOT", "github.com/silinternational/wecarry-api"))

		c.Set("rollbar", client)

		return next(c)
	}
}

func mergeExtras(extras []map[string]interface{}) map[string]interface{} {
	var allExtras map[string]interface{}

	// I didn't think I would need this, but without is at least one test was failing
	// The code allowed a map[string]interface{} to get through (i.e. not in a slice)
	// without the compiler complaining
	if len(extras) == 1 {
		return extras[0]
	}

	for _, e := range extras {
		for k, v := range e {
			allExtras[k] = v
		}
	}

	return allExtras
}

// Error log error and send to Rollbar
func Error(c buffalo.Context, msg string, extras ...map[string]interface{}) {
	// Avoid panics running tests when c doesn't have the necessary nested methods
	cType := fmt.Sprintf("%T", c)
	if cType == "models.EmptyContext" {
		return
	}

	es := mergeExtras(extras)
	c.Logger().Error(msg, es)
	rollbarMessage(c, rollbar.ERR, msg, es)
}

// Warn log warning and send to Rollbar
func Warn(c buffalo.Context, msg string, extras ...map[string]interface{}) {
	es := mergeExtras(extras)
	c.Logger().Warn(msg, es)
	rollbarMessage(c, rollbar.WARN, msg, es)
}

// Log info message
func Info(c buffalo.Context, msg string, extras ...map[string]interface{}) {
	c.Logger().Info(msg, mergeExtras(extras))
}

// rollbarMessage is a wrapper function to call rollbar's client.MessageWithExtras function from client stored in context
func rollbarMessage(c buffalo.Context, level string, msg string, extras map[string]interface{}) {
	rc, ok := c.Value("rollbar").(*rollbar.Client)
	if ok {
		rc.MessageWithExtras(level, msg, extras)
		return
	}
}

// RollbarSetPerson sets person on the rollbar context for futher logging
func RollbarSetPerson(c buffalo.Context, id, username, email string) {
	rc, ok := c.Value("rollbar").(*rollbar.Client)
	if ok {
		rc.SetPerson(id, username, email)
		return
	}
}
