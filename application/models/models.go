package models

import (
	"context"
	"crypto/md5" // #nosec G501 weak cryptography used for gravatar URL only
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
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
	if inString.Valid {
		output := inString.String
		return &output
	}

	return nil
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

// GetStringFromNullsTime returns a pointer to a string that looks
// like a date based on a nulls.Time value
func GetStringFromNullsTime(inTime nulls.Time) *string {
	if inTime.Valid {
		output := inTime.Time.Format(domain.DateFormat)
		return &output
	}

	return nil
}

// CurrentUser retrieves the current user from the context, which can be the context provided by gqlgen or the inner
// "BuffaloContext" assigned to the value key of the same name.
func CurrentUser(ctx context.Context) User {
	bc, ok := ctx.Value(domain.BuffaloContext).(buffalo.Context)
	if ok {
		return CurrentUser(bc)
	}
	user, _ := ctx.Value("current_user").(User)
	return user
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
	uuidField := fieldByName(m, "UUID")
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
	uuidField := fieldByName(m, "UUID")
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

func gravatarURL(email string) string {
	// ref: https://en.gravatar.com/site/implement/images/
	hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email)))) // #nosec G401 weak cryptography acceptable here
	url := fmt.Sprintf("https://www.gravatar.com/avatar/%x.jpg?s=200&d=mp", hash)
	return url
}

func addFile(m interface{}, fileID string) (File, error) {
	var f File

	if err := f.FindByUUID(fileID); err != nil {
		return f, err
	}

	fileField := fieldByName(m, "FileID")
	if !fileField.IsValid() {
		return f, errors.New("error identifying FileID field")
	}

	oldID := fileField.Interface().(nulls.Int)
	fileField.Set(reflect.ValueOf(nulls.NewInt(f.ID)))
	idField := fieldByName(m, "ID")
	if !idField.IsValid() {
		return f, errors.New("error identifying ID field")
	}
	if idField.Interface().(int) > 0 {
		if err := DB.UpdateColumns(m, "file_id"); err != nil {
			return f, fmt.Errorf("failed to update the file_id column, %s", err)
		}
	}

	if err := f.SetLinked(DB); err != nil {
		domain.ErrLogger.Printf("error marking file %d as linked, %s", f.ID, err)
	}

	if !oldID.Valid {
		return f, nil
	}

	oldFile := File{ID: oldID.Int}
	if err := oldFile.ClearLinked(DB); err != nil {
		domain.ErrLogger.Printf("error marking old file %d as unlinked, %s", oldFile.ID, err)
	}

	return f, nil
}

func removeFile(m interface{}) error {
	idField := fieldByName(m, "ID")
	if !idField.IsValid() {
		return errors.New("error identifying ID field")
	}

	if idField.Interface().(int) < 1 {
		return fmt.Errorf("invalid ID %d", idField.Interface().(int))
	}

	imageField := fieldByName(m, "FileID")
	if !imageField.IsValid() {
		return errors.New("error identifying FileID field")
	}

	oldID := imageField.Interface().(nulls.Int)
	imageField.Set(reflect.ValueOf(nulls.Int{}))
	if err := DB.UpdateColumns(m, "file_id"); err != nil {
		return fmt.Errorf("failed to update file_id column, %s", err)
	}

	if !oldID.Valid {
		return nil
	}

	oldFile := File{ID: oldID.Int}
	if err := oldFile.ClearLinked(DB); err != nil {
		domain.ErrLogger.Printf("error marking old meeting file %d as unlinked, %s", oldFile.ID, err)
	}
	return nil
}

func fieldByName(i interface{}, name string) reflect.Value {
	return reflect.ValueOf(i).Elem().FieldByName(name)
}
