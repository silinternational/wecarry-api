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
	ErrorLevelWarn             = "warn"
	ErrorLevelError            = "error"
	ErrorLevelCritical         = "critical"
	AdminRoleSuperDuperAdmin   = "SuperDuperAdmin"
	AdminRoleSalesAdmin        = "SalesAdmin"
	EmptyUUID                  = "00000000-0000-0000-0000-000000000000"
	DateFormat                 = "2006-01-02"
	MaxFileSize                = 1 << 20 // 1 Mebibyte
	AccessTokenLifetimeSeconds = 3600
	DateTimeFormat             = "2006-01-02 15:04:05"

	// Environment Variables
	UIURLEnv                      = "UI_URL"
	AccessTokenLifetimeSecondsEnv = "ACCESS_TOKEN_LIFETIME_SECONDS"
)

// NoExtras is exported for use when making calls to RollbarError and rollbarMessage to reduce
// typing map[string]interface{} when no extras are needed
var NoExtras map[string]interface{}

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

// GetRequestData parses the URL, if the method is GET, or the body, if the method
// is POST or PUT, and returns a map[string][]string with all of the parameter/value
// pairs. In either case, the data must be urlencoded.
func GetRequestData(r *http.Request) (map[string][]string, error) {
	data := map[string][]string{}

	if r.Method == "GET" {
		return r.URL.Query(), nil
	}

	if r.Method == "POST" || r.Method == "PUT" {
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			return data, fmt.Errorf("error getting POST data: %v", err.Error())
		}

		data = r.PostForm
	}

	return data, nil
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
	// TODO: Handle this error
	uuid, _ := uuid2.NewV4()
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

// Error log error and send to Rollbar
func Error(c buffalo.Context, msg string, extras map[string]interface{}) {
	c.Logger().Error(msg, extras)
	rollbarMessage(c, rollbar.ERR, msg, extras)
}

// Warn log warning and send to Rollbar
func Warn(c buffalo.Context, msg string, extras map[string]interface{}) {
	c.Logger().Warn(msg, extras)
	rollbarMessage(c, rollbar.WARN, msg, extras)
}

// Log info message
func Info(c buffalo.Context, msg string, extras map[string]interface{}) {
	c.Logger().Info(msg, extras)
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
