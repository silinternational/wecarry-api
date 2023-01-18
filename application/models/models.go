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

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
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

// GetIntFromNullsInt returns a pointer to make it easier for calling
// functions to return a pointer without an extra line of code.
func GetIntFromNullsInt(in nulls.Int) *int {
	output := 0
	if in.Valid {
		output = in.Int
	}
	return &output
}

// CurrentUser retrieves the current user from the context.
func CurrentUser(ctx context.Context) User {
	user, _ := ctx.Value(domain.ContextKeyCurrentUser).(User)
	domain.NewExtra(ctx, "currentUserID", user.UUID)
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

func create(tx *pop.Connection, m interface{}) error {
	uuidField := fieldByName(m, "UUID")
	if uuidField.IsValid() && uuidField.Interface().(uuid.UUID).Version() == 0 {
		uuidField.Set(reflect.ValueOf(domain.GetUUID()))
	}

	valErrs, err := tx.ValidateAndCreate(m)
	if err != nil {
		return err
	}

	if valErrs.HasAny() {
		return errors.New(flattenPopErrors(valErrs))
	}
	return nil
}

func update(tx *pop.Connection, m interface{}) error {
	valErrs, err := tx.ValidateAndUpdate(m)
	if err != nil {
		return err
	}

	if valErrs.HasAny() {
		return errors.New(flattenPopErrors(valErrs))
	}
	return nil
}

func save(tx *pop.Connection, m interface{}) error {
	uuidField := fieldByName(m, "UUID")
	if uuidField.IsValid() && uuidField.Interface().(uuid.UUID).Version() == 0 {
		uuidField.Set(reflect.ValueOf(domain.GetUUID()))
	}

	validationErrs, err := tx.ValidateAndSave(m)
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

func addFile(tx *pop.Connection, m interface{}, fileID string) (File, error) {
	var f File

	if err := f.FindByUUID(tx, fileID); err != nil {
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
		if err := tx.UpdateColumns(m, "file_id", "updated_at"); err != nil {
			return f, fmt.Errorf("failed to update the file_id column, %s", err)
		}
	}

	if err := f.SetLinked(tx); err != nil {
		domain.ErrLogger.Printf("error marking file %d as linked, %s", f.ID, err)
	}

	if !oldID.Valid {
		return f, nil
	}

	oldFile := File{ID: oldID.Int}
	if err := oldFile.ClearLinked(tx); err != nil {
		domain.ErrLogger.Printf("error marking old file %d as unlinked, %s", oldFile.ID, err)
	}

	return f, nil
}

func removeFile(tx *pop.Connection, m interface{}) error {
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
	if err := tx.UpdateColumns(m, "file_id", "updated_at"); err != nil {
		return fmt.Errorf("failed to update file_id column, %s", err)
	}

	if !oldID.Valid {
		return nil
	}

	oldFile := File{ID: oldID.Int}
	if err := oldFile.ClearLinked(tx); err != nil {
		domain.ErrLogger.Printf("error marking old meeting file %d as unlinked, %s", oldFile.ID, err)
	}
	return nil
}

func fieldByName(i interface{}, name string) reflect.Value {
	return reflect.ValueOf(i).Elem().FieldByName(name)
}

func DestroyAll() {
	// delete all Requests, RequestHistories, RequestFiles, PotentialProviders, Threads, and ThreadParticipants
	var requests Requests
	destroyTable(&requests)

	// delete all Meetings, MeetingParticipants, and MeetingInvites
	var meetings Meetings
	destroyTable(&meetings)

	// delete all Organizations, OrganizationDomains, OrganizationTrusts, and UserOrganizations
	var organizations Organizations
	destroyTable(&organizations)

	// delete all Users, Messages, UserAccessTokens, and Watches
	var users Users
	destroyTable(&users)

	// delete all Files
	var files Files
	destroyTable(&files)

	// delete all Locations
	var locations Locations
	destroyTable(&locations)
}

func destroyTable(i interface{}) {
	if err := DB.All(i); err != nil {
		panic(err.Error())
	}
	if err := DB.Destroy(i); err != nil {
		panic(err.Error())
	}
}

// Tx retrieves the database transaction from the context
func Tx(ctx context.Context) *pop.Connection {
	tx, ok := ctx.Value(domain.ContextKeyTx).(*pop.Connection)
	if !ok {
		domain.Logger.Print("no transaction found in context, called from: " + domain.GetFunctionName(2))
		return DB
	}
	return tx
}
