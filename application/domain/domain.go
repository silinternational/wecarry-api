package domain

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	i18n "github.com/gobuffalo/mw-i18n"
	uuid2 "github.com/gofrs/uuid"
	"github.com/rollbar/rollbar-go"
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

// Event Kinds
const (
	EventApiUserCreated       = "api:user:created"
	EventApiAuthUserLoggedIn  = "api:auth:user:loggedin"
	EventApiMessageCreated    = "api:message:created"
	EventApiPostStatusUpdated = "api:post:status:updated"
)

// Event and Job argument names
const (
	ArgMessageID = "message_id"
)

// Notification Message Template Names
const (
	MessageTemplateNewMessage                      = "new_message"
	MessageTemplateNewRequest                      = "new_request"
	MessageTemplateRequestFromCommittedToOpen      = "request_from_committed_to_open"
	MessageTemplateRequestFromAcceptedToOpen       = "request_from_accepted_to_open"
	MessageTemplateRequestFromOpenToCommitted      = "request_from_open_to_committed"
	MessageTemplateRequestFromCommittedToAccepted  = "request_from_committed_to_accepted"
	MessageTemplateRequestFromDeliveredToAccepted  = "request_from_delivered_to_accepted"
	MessageTemplateRequestFromReceivedToAccepted   = "request_from_received_to_accepted"
	MessageTemplateRequestFromCommittedToDelivered = "request_from_committed_to_delivered"
	MessageTemplateRequestFromAcceptedToDelivered  = "request_from_accepted_to_delivered"
	MessageTemplateRequestFromReceivedToDelivered  = "request_from_received_to_delivered"
	MessageTemplateRequestFromCompletedToDelivered = "request_from_completed_to_delivered"
	MessageTemplateRequestFromAcceptedToReceived   = "request_from_accepted_to_received"
	MessageTemplateRequestFromCompletedToReceived  = "request_from_completed_to_received"
	MessageTemplateRequestFromDeliveredToCompleted = "request_from_delivered_to_completed"
	MessageTemplateRequestFromReceivedToCompleted  = "request_from_received_to_completed"
	MessageTemplateRequestFromOpenToRemoved        = "request_from_open_to_removed"
	MessageTemplateRequestFromCommittedToRemoved   = "request_from_committed_to_removed"
	MessageTemplateRequestFromAcceptedToRemoved    = "request_from_accepted_to_removed"
)

// UI URL Paths
const (
	postUIPath   = "/#/requests/"
	threadUIPath = "/#/messages/"
)

var Logger log.Logger
var ErrLogger log.Logger

// Env holds environment variable values loaded by init()
var Env struct {
	AccessTokenLifetimeSeconds int
	AuthCallbackURL            string
	AwsS3Region                string
	AwsS3Endpoint              string
	AwsS3DisableSSL            bool
	AwsS3Bucket                string
	AwsS3AccessKeyID           string
	AwsS3SecretAccessKey       string
	EmailService               string
	GoEnv                      string
	GoogleKey                  string
	GoogleSecret               string
	MobileService              string
	PlaygroundPort             string
	RollbarServerRoot          string
	RollbarToken               string
	SendGridAPIKey             string
	SessionSecret              string
	UIURL                      string
}

// T is the Buffalo i18n translator
var T *i18n.Translator

func init() {
	Logger.SetOutput(os.Stdout)
	ErrLogger.SetOutput(os.Stderr)

	ReadEnv()
}

// ReadEnv loads environment data into `Env`
func ReadEnv() {
	n, err := strconv.Atoi(envy.Get("ACCESS_TOKEN_LIFETIME_SECONDS", strconv.Itoa(AccessTokenLifetimeSeconds)))
	if err != nil {
		ErrLogger.Printf("error converting token lifetime env var ... %v", err)
		n = AccessTokenLifetimeSeconds
	}
	Env.AccessTokenLifetimeSeconds = n
	Env.AuthCallbackURL = envy.Get("AUTH_CALLBACK_URL", "")
	Env.AwsS3Region = envy.Get("AWS_REGION", "")
	Env.AwsS3Endpoint = envy.Get("AWS_S3_ENDPOINT", "")
	Env.AwsS3DisableSSL, _ = strconv.ParseBool(envy.Get("AWS_S3_DISABLE_SSL", "false"))
	Env.AwsS3Bucket = envy.Get("AWS_S3_BUCKET", "")
	Env.AwsS3AccessKeyID = envy.Get("AWS_S3_ACCESS_KEY_ID", "")
	Env.AwsS3SecretAccessKey = envy.Get("AWS_S3_SECRET_ACCESS_KEY", "")
	Env.EmailService = envy.Get("EMAIL_SERVICE", "sendgrid")
	Env.GoEnv = envy.Get("GO_ENV", "development")
	Env.GoogleKey = envy.Get("GOOGLE_KEY", "")
	Env.GoogleSecret = envy.Get("GOOGLE_SECRET", "")
	Env.MobileService = envy.Get("MOBILE_SERVICE", "dummy")
	Env.PlaygroundPort = envy.Get("PORT", "3000")
	Env.RollbarServerRoot = envy.Get("ROLLBAR_SERVER_ROOT", "github.com/silinternational/wecarry-api")
	Env.RollbarToken = envy.Get("ROLLBAR_TOKEN", "")
	Env.SendGridAPIKey = envy.Get("SENDGRID_API_KEY", "")
	Env.SessionSecret = envy.Get("SESSION_SECRET", "testing")
	Env.UIURL = envy.Get("UI_URL", "dev.wecarry.app")
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
			Env.RollbarToken,
			Env.GoEnv,
			"",
			"",
			Env.RollbarServerRoot)

		c.Set("rollbar", client)

		return next(c)
	}
}

func mergeExtras(extras []map[string]interface{}) map[string]interface{} {
	var allExtras map[string]interface{}

	// I didn't think I would need this, but without it at least one test was failing
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

// GetPostUIURL returns a UI URL for the given Post
func GetPostUIURL(postUUID string) string {
	return Env.UIURL + postUIPath + postUUID
}

// GetThreadUIURL returns a UI URL for the given Thread
func GetThreadUIURL(threadUUID string) string {
	return Env.UIURL + threadUIPath + threadUUID
}
