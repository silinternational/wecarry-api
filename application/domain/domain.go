package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	mwi18n "github.com/gobuffalo/mw-i18n"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	uuid2 "github.com/gofrs/uuid"
	"github.com/rollbar/rollbar-go"
)

const (
	Megabyte                    = 1048576
	DateFormat                  = "2006-01-02"
	MaxFileSize                 = 1024 * 1024 * 10       // 10 Megabytes
	AccessTokenLifetimeSeconds  = 60*60*24*13 + 60*60*12 // 13 days, 12 hours
	DateTimeFormat              = "2006-01-02 15:04:05"
	NewMessageNotificationDelay = 1 * time.Minute
	DefaultProximityDistanceKm  = 100
	DurationDay                 = time.Duration(time.Hour * 24)
	DurationWeek                = time.Duration(DurationDay * 7)
	RecentMeetingDelay          = DurationDay * 30
)

// Event Kinds
const (
	EventApiUserCreated                    = "api:user:created"
	EventApiAuthUserLoggedIn               = "api:auth:user:loggedin"
	EventApiMessageCreated                 = "api:message:created"
	EventApiPostStatusUpdated              = "api:post:status:updated"
	EventApiPostCreated                    = "api:post:status:created"
	EventApiPotentialProviderCreated       = "api:potentialprovider:created"
	EventApiPotentialProviderRejected      = "api:potentialprovider:rejected"
	EventApiPotentialProviderSelfDestroyed = "api:potentialprovider:selfdestroyed"
)

// Event and Job argument names
const (
	ArgMessageID = "message_id"
)

// Notification Message Template Names
const (
	MessageTemplateNewRequest                      = "new_request"
	MessageTemplateNewThreadMessage                = "new_thread_message"
	MessageTemplateNewUserWelcome                  = "new_user_welcome"
	MessageTemplateRequestFromAcceptedToCompleted  = "request_from_accepted_to_completed"
	MessageTemplateRequestFromAcceptedToDelivered  = "request_from_accepted_to_delivered"
	MessageTemplateRequestFromAcceptedToOpen       = "request_from_accepted_to_open"
	MessageTemplateRequestFromAcceptedToReceived   = "request_from_accepted_to_received"
	MessageTemplateRequestFromAcceptedToRemoved    = "request_from_accepted_to_removed"
	MessageTemplateRequestFromCompletedToAccepted  = "request_from_completed_to_accepted"
	MessageTemplateRequestFromCompletedToDelivered = "request_from_completed_to_delivered"
	MessageTemplateRequestFromCompletedToReceived  = "request_from_completed_to_received"
	MessageTemplateRequestFromDeliveredToAccepted  = "request_from_delivered_to_accepted"
	MessageTemplateRequestFromDeliveredToCompleted = "request_from_delivered_to_completed"
	MessageTemplateRequestFromOpenToAccepted       = "request_from_open_to_accepted"
	MessageTemplateRequestFromOpenToRemoved        = "request_from_open_to_removed"
	MessageTemplateRequestFromReceivedToCompleted  = "request_from_received_to_completed"
	MessageTemplateRequestDelivered                = "request_delivered"
	MessageTemplateRequestReceived                 = "request_received"
	MessageTemplateRequestNotReceivedAfterAll      = "request_not_received_after_all"
	MessageTemplatePotentialProviderCreated        = "request_potentialprovider_created"
	MessageTemplatePotentialProviderRejected       = "request_potentialprovider_rejected"
	MessageTemplatePotentialProviderSelfDestroyed  = "request_potentialprovider_self_destroyed"
)

// User preferences
const (
	UserPreferenceKeyLanguage        = "language"
	UserPreferenceLanguageEnglish    = "en"
	UserPreferenceLanguageFrench     = "fr"
	UserPreferenceLanguageSpanish    = "es"
	UserPreferenceLanguageKorean     = "ko"
	UserPreferenceLanguagePortuguese = "pt"

	UserPreferenceKeyTimeZone = "time_zone"

	UserPreferenceKeyWeightUnit    = "weight_unit"
	UserPreferenceWeightUnitPounds = "pounds"
	UserPreferenceWeightUnitKGs    = "kilograms"
)

// UI URL Paths
const (
	DefaultUIPath = "/#/requests"
	postUIPath    = "/#/requests/"
	threadUIPath  = "/#/messages/"
)

// BuffaloContextType is a custom type used as a value key passed to context.WithValue as per the recommendations
// in the function docs for that function: https://golang.org/pkg/context/#WithValue
type BuffaloContextType string

// BuffaloContext is the key for the call to context.WithValue in gqlHandler
const BuffaloContext = BuffaloContextType("BuffaloContext")

var Logger log.Logger
var ErrLogger ErrLogProxy
var AuthCallbackURL string

// Env holds environment variable values loaded by init()
var Env struct {
	AccessTokenLifetimeSeconds int
	ServiceIntegrationToken    string
	ApiBaseURL                 string
	AppName                    string
	AuthCallbackURL            string
	AwsRegion                  string
	AwsS3Endpoint              string
	AwsS3DisableSSL            bool
	AwsS3Bucket                string
	AwsAccessKeyID             string
	AwsSecretAccessKey         string
	CertDomainName             string
	CloudflareAuthEmail        string
	CloudflareAuthKey          string
	DisableTLS                 bool
	DynamoDBTable              string
	EmailService               string
	EmailFromAddress           string
	FacebookKey                string
	FacebookSecret             string
	GoEnv                      string
	GoogleKey                  string
	GoogleSecret               string
	LinkedInKey                string
	LinkedInSecret             string
	MaxFileDelete              int
	MailChimpAPIBaseURL        string
	MailChimpAPIKey            string
	MailChimpListID            string
	MailChimpUsername          string
	MicrosoftKey               string
	MicrosoftSecret            string
	MobileService              string
	PlaygroundPort             string
	RollbarServerRoot          string
	RollbarToken               string
	SendGridAPIKey             string
	ServerPort                 int
	SessionSecret              string
	SupportEmail               string
	TwitterKey                 string
	TwitterSecret              string
	UIURL                      string
}

// T is the Buffalo i18n translator
var T *mwi18n.Translator

// Assets is a packr box with asset files such as images
var Assets *packr.Box

func init() {
	readEnv()
	Logger.SetOutput(os.Stdout)
	ErrLogger.SetOutput(os.Stderr)
	ErrLogger.InitRollbar()
	Assets = packr.New("Assets", "../assets")
	AuthCallbackURL = Env.ApiBaseURL + "/auth/callback"
}

// readEnv loads environment data into `Env`
func readEnv() {
	Env.AccessTokenLifetimeSeconds = envToInt("ACCESS_TOKEN_LIFETIME_SECONDS", AccessTokenLifetimeSeconds)
	Env.ApiBaseURL = envy.Get("HOST", "")
	Env.AppName = envy.Get("APP_NAME", "WeCarry")
	Env.AuthCallbackURL = envy.Get("AUTH_CALLBACK_URL", "")
	Env.AwsRegion = envy.Get("AWS_DEFAULT_REGION", "")
	Env.AwsS3Endpoint = envy.Get("AWS_S3_ENDPOINT", "")
	Env.AwsS3DisableSSL, _ = strconv.ParseBool(envy.Get("AWS_S3_DISABLE_SSL", "false"))
	Env.AwsS3Bucket = envy.Get("AWS_S3_BUCKET", "")
	Env.AwsAccessKeyID = envy.Get("AWS_ACCESS_KEY_ID", "")
	Env.AwsSecretAccessKey = envy.Get("AWS_SECRET_ACCESS_KEY", "")
	Env.CertDomainName = envy.Get("CERT_DOMAIN_NAME", "")
	Env.CloudflareAuthEmail = envy.Get("CLOUDFLARE_AUTH_EMAIL", "")
	Env.CloudflareAuthKey = envy.Get("CLOUDFLARE_AUTH_KEY", "")
	Env.DisableTLS, _ = strconv.ParseBool(envy.Get("DISABLE_TLS", "false"))
	Env.DynamoDBTable = envy.Get("DYNAMO_DB_TABLE", "CertMagic")
	Env.EmailService = envy.Get("EMAIL_SERVICE", "sendgrid")
	Env.EmailFromAddress = envy.Get("EMAIL_FROM_ADDRESS", "no_reply@example.com")
	Env.FacebookKey = envy.Get("FACEBOOK_KEY", "")
	Env.FacebookSecret = envy.Get("FACEBOOK_SECRET", "")
	Env.GoEnv = envy.Get("GO_ENV", "development")
	Env.GoogleKey = envy.Get("GOOGLE_KEY", "")
	Env.GoogleSecret = envy.Get("GOOGLE_SECRET", "")
	Env.LinkedInKey = envy.Get("LINKED_IN_KEY", "")
	Env.LinkedInSecret = envy.Get("LINKED_IN_SECRET", "")
	Env.MaxFileDelete = envToInt("MAX_FILE_DELETE", 10)
	Env.MailChimpAPIBaseURL = envy.Get("MAILCHIMP_API_BASE_URL", "https://us4.api.mailchimp.com/3.0")
	Env.MailChimpAPIKey = envy.Get("MAILCHIMP_API_KEY", "")
	Env.MailChimpListID = envy.Get("MAILCHIMP_LIST_ID", "")
	Env.MailChimpUsername = envy.Get("MAILCHIMP_USERNAME", "")
	Env.MicrosoftKey = envy.Get("MICROSOFT_KEY", "")
	Env.MicrosoftSecret = envy.Get("MICROSOFT_SECRET", "")
	Env.MobileService = envy.Get("MOBILE_SERVICE", "dummy")
	Env.PlaygroundPort = envy.Get("PORT", "3000")
	Env.RollbarServerRoot = envy.Get("ROLLBAR_SERVER_ROOT", "github.com/silinternational/wecarry-api")
	Env.RollbarToken = envy.Get("ROLLBAR_TOKEN", "")
	Env.SendGridAPIKey = envy.Get("SENDGRID_API_KEY", "")
	Env.ServerPort, _ = strconv.Atoi(envy.Get("PORT", "3000"))
	Env.ServiceIntegrationToken = envy.Get("SERVICE_INTEGRATION_TOKEN", "")
	Env.SessionSecret = envy.Get("SESSION_SECRET", "testing")
	Env.SupportEmail = envy.Get("SUPPORT_EMAIL", "")
	Env.TwitterKey = envy.Get("TWITTER_KEY", "")
	Env.TwitterSecret = envy.Get("TWITTER_SECRET", "")
	Env.UIURL = envy.Get("UI_URL", "dev.wecarry.app")
}

func envToInt(name string, def int) int {
	s := envy.Get(name, strconv.Itoa(def))
	n, err := strconv.Atoi(s)
	if err != nil {
		ErrLogger.Printf("invalid environment variable %s = %s, must be a number, %s", name, s, err)
		return def
	}
	return n
}

type AppError struct {
	Code int `json:"code"`

	// Don't change the value of these Key entries without making a corresponding change on the UI,
	// since these will be converted to human-friendly texts on the UI
	Key string `json:"key"`
}

// ErrLogProxy wraps standard error logger plus sends to Rollbar
type ErrLogProxy struct {
	LocalLog  log.Logger
	RemoteLog *rollbar.Client
}

func (e *ErrLogProxy) SetOutput(w io.Writer) {
	e.LocalLog.SetOutput(w)
}

func (e *ErrLogProxy) Printf(format string, a ...interface{}) {
	// Send to local logger
	e.LocalLog.Printf(format, a...)

	// Only send to remote log if not in test env
	if Env.GoEnv == "test" {
		return
	}
	e.RemoteLog.Errorf(rollbar.ERR, format, a...)
}

func (e *ErrLogProxy) InitRollbar() {
	e.RemoteLog = rollbar.New(
		Env.RollbarToken,
		Env.GoEnv,
		"",
		"",
		Env.RollbarServerRoot)
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

// GetUUID creates a new, unique version 4 (random) UUID and returns it
// as a uuid2.UUID. Errors are ignored.
func GetUUID() uuid2.UUID {
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
		return strings.ToLower(parts[1])
	} else {
		return strings.ToLower(email)
	}
}

func RollbarMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if Env.RollbarToken == "" {
			return next(c)
		}

		if Env.GoEnv == "test" {
			return next(c)
		}

		client := rollbar.New(
			Env.RollbarToken,
			Env.GoEnv,
			"",
			"",
			Env.RollbarServerRoot)

		c.Set("rollbar", client)

		err := next(c)

		client.Close()
		return err
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
	logger := c.Logger()
	if logger == nil {
		return
	}

	es := mergeExtras(extras)
	logger.Error(msg, es)
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

// RollbarSetPerson sets person on the rollbar context for further logging
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

func IsLanguageAllowed(lang string) bool {
	switch lang {
	case UserPreferenceLanguageEnglish, UserPreferenceLanguageFrench, UserPreferenceLanguageKorean,
		UserPreferenceLanguagePortuguese, UserPreferenceLanguageSpanish:
		return true
	}

	return false
}

func IsWeightUnitAllowed(unit string) bool {
	switch unit {
	case UserPreferenceWeightUnitKGs, UserPreferenceWeightUnitPounds:
		return true
	}

	return false
}

func IsTimeZoneAllowed(name string) bool {
	_, err := time.LoadLocation(name)

	if err != nil {
		Logger.Printf("error evaluating timezone %s ... %v", name, err)
		return false
	}

	return true
}

// TranslateWithLang returns the translation of the string identified by translationID, for the given language.
// Apparently i18n has a global or something that keeps track of translatable phrases once a new packr Box
// is created.  If no new packr Box has been created, i18n.Tfunc returns an error.
func TranslateWithLang(lang, translationID string, args ...interface{}) (string, error) {
	if T == nil {
		_, err := mwi18n.New(packr.New("locales", "../locales"), "en")
		if err != nil {
			return "", err
		}
	}

	return T.TranslateWithLang(lang, translationID, args...)
}

// IsOtherThanNoRows returns false if the error is nil or is just reporting that there
//   were no rows in the result set for a sql query.
func IsOtherThanNoRows(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), "sql: no rows in result set") {
		return false
	}

	return true
}

// GetStructTags creates a map with certain types of tags (e.g. json) of a struct's
// fields.  That tag values are both the keys and values of the map - just to make
// it easy to check if a certain value is in the map
func GetStructTags(tagType string, s interface{}) (map[string]string, error) {
	rt := reflect.TypeOf(s)
	if rt.Kind() != reflect.Struct {
		return map[string]string{}, fmt.Errorf("cannot get fieldTags of non structs, not even for %v", rt.Kind())
	}

	fieldTags := map[string]string{}

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)

		v := strings.Split(f.Tag.Get(tagType), ",")[0] // use split to ignore tag "options" like omitempty, etc.
		if v != "" {
			fieldTags[v] = v
		}
	}

	return fieldTags, nil
}

func GetTranslatedSubject(language, translationID string, translationData map[string]string) string {
	translationData["AppName"] = Env.AppName

	subj, err := TranslateWithLang(language, translationID, translationData)

	if err != nil {
		ErrLogger.Printf("error translating '%s' notification subject, %s", translationID, err)
	}

	return subj
}

// Truncate returns the given string truncated to length including the suffix if originally longer than length
func Truncate(str, suffix string, length int) string {
	a := []rune(str)
	s := []rune(suffix)
	if len(a) > length {
		return string(a[0:length-len(s)]) + suffix
	}
	return str
}

// EmailFromAddress combines a name with the configured from address for use in an email From header. If name is nil,
// only the App Name will be used.
func EmailFromAddress(name *string) string {
	addr := Env.AppName + " <" + Env.EmailFromAddress + ">"
	if name != nil {
		addr = *name + " via " + addr
	}
	return addr
}

// RemoveUnwantedChars removes characters from `str` that are not in `allowed` and not in "safe" character ranges.
func RemoveUnwantedChars(str, allowed string) string {
	filter := func(r rune) rune {
		if strings.IndexRune(allowed, r) >= 0 || isSafeRune(r) {
			return r
		}
		return -1
	}
	return strings.Map(filter, str)
}

func isSafeRune(r rune) bool {
	var safeRanges = &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0x0030, Hi: 0x0039, Stride: 1}, // 0-9
			{Lo: 0x0041, Hi: 0x005a, Stride: 1}, // upper-case Latin
			{Lo: 0x0061, Hi: 0x007a, Stride: 1}, // lower-case Latin
			{Lo: 0x00c0, Hi: 0x00ff, Stride: 1}, // Latin 1 supplement
			{Lo: 0x0100, Hi: 0x017f, Stride: 1}, // Latin extended A
			{Lo: 0x200d, Hi: 0x200d, Stride: 1}, // Zero-width Joiner
			{Lo: 0x2600, Hi: 0x26ff, Stride: 1}, // symbols
			{Lo: 0x2700, Hi: 0x27bf, Stride: 1}, // dingbat
			{Lo: 0xac00, Hi: 0xd7a3, Stride: 1}, // Hangul (Korean)
			{Lo: 0xfe0f, Hi: 0xfe0f, Stride: 1}, // variation selector 16
		},
		R32: []unicode.Range32{
			{Lo: 0x1f1e6, Hi: 0x1f1ff, Stride: 1}, // regional indicator symbol
			{Lo: 0x1f300, Hi: 0x1f5ff, Stride: 1}, // pictographs
			{Lo: 0x1f600, Hi: 0x1f64f, Stride: 1}, // emoji
			{Lo: 0x1f680, Hi: 0x1f6ff, Stride: 1}, // transport
			{Lo: 0x1f900, Hi: 0x1f9ff, Stride: 1}, // supplemental symbols and pictographs
		},
	}

	return unicode.In(r, safeRanges)
}

// StringIsVisible is a validator to ensure at least one character is visible, i.e. not whitespace or control char.
type StringIsVisible struct {
	Name    string
	Field   string
	Message string
}

// IsValid adds an error if the field is not empty and not a url.
func (v *StringIsVisible) IsValid(errors *validate.Errors) {
	var asciiSpace = map[int32]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true}
	for _, c := range v.Field {
		if !asciiSpace[c] && unicode.IsGraphic(c) && c != 0xfe0f { // VS16 (0xfe0f) is "Graphic" but invisible
			return
		}
	}

	if len(v.Message) > 0 {
		errors.Add(validators.GenerateKey(v.Name), v.Message)
		return
	}

	errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("%s must have a visible character.", v.Name))
}

// ReportError logs an error with details, and returns a user-friendly, translated error identified by translation key
// string `errID`. If called with a full GraphQL context, the query text will be logged in the extras.
func ReportError(ctx context.Context, err error, errID string, extras ...map[string]interface{}) error {
	c := GetBuffaloContext(ctx)
	allExtras := map[string]interface{}{
		"function": GetFunctionName(2),
	}
	if r := graphql.GetRequestContext(ctx); r != nil {
		allExtras["query"] = r.RawQuery
	}
	for _, e := range extras {
		for key, val := range e {
			allExtras[key] = val
		}
	}

	errStr := errID
	if err != nil {
		errStr = err.Error()
	}
	Error(c, errStr, allExtras)

	if T == nil {
		return errors.New(errID)
	}
	return errors.New(T.Translate(c, errID))
}

// GetBuffaloContext retrieves a "BuffaloContext" from a wrapped context as constructed by
// actions.gqlHandler. If it's already a buffalo.Context, it is returned as is, type casted to buffalo.Context.
func GetBuffaloContext(c context.Context) buffalo.Context {
	bc, ok := c.Value(BuffaloContext).(buffalo.Context)
	if ok {
		return bc
	}
	bc, ok = c.(buffalo.Context)
	if ok {
		return bc
	}
	return emptyContext{}
}

type emptyContext struct {
	buffalo.Context
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
