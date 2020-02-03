package models

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/silinternational/wecarry-api/domain"
)

// Count can be used to receive the results of a SQL COUNT
type Count struct {
	N int `db:"count"`
}

// DB is a connection to the database to be used throughout the application.
var DB *pop.Connection

const tokenBytes = 32

// Keep a map of the json tag names for the standard user preferences struct
// e.g. "time_zone": "time_zone".
// Having it as a map, makes it easy to check if a potential key is allowed
var allowedUserPreferenceKeys map[string]string

func init() {
	var err error
	env := domain.Env.GoEnv
	DB, err = pop.Connect(env)
	if err != nil {
		domain.ErrLogger.Printf("error connecting to database ... %v", err)
		log.Fatal(err)
	}
	pop.Debug = env == "development"

	// Just make sure we can use the crypto/rand library on our system
	if _, err = getRandomToken(); err != nil {
		log.Fatal(fmt.Errorf("error using crypto/rand ... %v", err))
	}

	allowedUserPreferenceKeys, err = domain.GetStructTags("json", StandardPreferences{})
	if err != nil {
		log.Fatal(fmt.Errorf("error loading Allowed User Preferences ... %v", err))
	}
}

func getRandomToken() (string, error) {
	rb := make([]byte, tokenBytes)

	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(rb), nil
}

func ConvertStringPtrToNullsString(inPtr *string) nulls.String {
	if inPtr == nil {
		return nulls.String{}
	}

	return nulls.NewString(*inPtr)
}

// GetStringFromNullsString returns a pointer to make it easier for calling
// functions to return a pointer without an extra line of code.
func GetStringFromNullsString(inString nulls.String) *string {
	output := ""
	if inString.Valid {
		output = inString.String
	}

	return &output
}

// GetIntFromNullsInt returns a pointer to make it easier for calling
// functions to return a pointer without an extra line of code.
func GetIntFromNullsInt(in nulls.Int) *int {
	output := int(0)
	if in.Valid {
		output = in.Int
	}
	return &output
}

func GetCurrentUserFromGqlContext(ctx context.Context) User {
	bc, ok := ctx.Value("BuffaloContext").(buffalo.Context)
	if !ok {
		return User{}
	}
	return GetCurrentUser(bc)
}

type EmptyContext struct {
	buffalo.Context
}

func GetBuffaloContextFromGqlContext(c context.Context) buffalo.Context {
	bc, ok := c.Value("BuffaloContext").(buffalo.Context)
	if ok {
		return bc
	}
	return EmptyContext{}
}

func GetCurrentUser(c buffalo.Context) User {
	user := c.Value("current_user")

	switch user.(type) {
	case User:
		return user.(User)
	}

	return User{}
}

// flattenPopErrors - pop validation errors are complex structures, this flattens them to a simple string
func flattenPopErrors(popErrs *validate.Errors) string {
	var msg string
	for key, val := range popErrs.Errors {
		msg += fmt.Sprintf("%s: %s |", key, strings.Join(val, ", "))
	}

	return msg
}

// NullsStringIsURL is a model field validator
// which makes sure that a NullsString that is not blank or null is
// a valid URL
type NullsStringIsURL struct {
	Name    string
	Field   nulls.String
	Message string
}

// IsValid adds an error if the field is not empty and not a url.
func (v *NullsStringIsURL) IsValid(errors *validate.Errors) {
	if !v.Field.Valid {
		return
	}
	value := v.Field.String

	if value == "" {
		return
	}

	newV := validators.URLIsPresent{Name: v.Name, Field: value, Message: v.Message}
	newV.IsValid(errors)
}

// This can include an event payload, which is a map[string]interface{}
func emitEvent(e events.Event) {
	if err := events.Emit(e); err != nil {
		domain.ErrLogger.Printf("error emitting event %s ... %v", e.Kind, err)
	}
}

func create(m interface{}) error {
	uuidField := reflect.ValueOf(m).Elem().FieldByName("UUID")
	if uuidField.IsValid() && uuidField.Interface().(uuid.UUID).Version() == 0 {
		uuidField.Set(reflect.ValueOf(domain.GetUUID()))
	}

	valErrs, err := DB.ValidateAndCreate(m)
	if err != nil {
		return err
	}

	if valErrs.HasAny() {
		return errors.New(flattenPopErrors(valErrs))
	}
	return nil
}

func update(m interface{}) error {
	valErrs, err := DB.ValidateAndUpdate(m)
	if err != nil {
		return err
	}

	if valErrs.HasAny() {
		return errors.New(flattenPopErrors(valErrs))
	}
	return nil
}

func save(m interface{}) error {
	uuidField := reflect.ValueOf(m).Elem().FieldByName("UUID")
	if uuidField.IsValid() && uuidField.Interface().(uuid.UUID).Version() == 0 {
		uuidField.Set(reflect.ValueOf(domain.GetUUID()))
	}

	validationErrs, err := DB.ValidateAndSave(m)
	if validationErrs != nil && validationErrs.HasAny() {
		return errors.New(flattenPopErrors(validationErrs))
	}
	if err != nil {
		return err
	}
	return nil
}

func convertSliceFromIntToInterface(intSlice []int) []interface{} {
	s := make([]interface{}, len(intSlice))
	for i, v := range intSlice {
		s[i] = v
	}
	return s
}

func IsDBConnected() bool {
	var org Organization
	if err := DB.First(&org); err != nil {
		return !domain.IsOtherThanNoRows(err)
	}
	return true
}
