package models

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gobuffalo/validate/validators"
	"log"
	"strings"

	"github.com/pkg/errors"

	"github.com/gobuffalo/validate"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection

func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	DB, err = pop.Connect(env)
	if err != nil {
		log.Fatal(err)
	}
	pop.Debug = env == "development"
}

func ConvertStringPtrToNullsString(inPtr *string) nulls.String {
	if inPtr == nil {
		return nulls.String{}
	}

	return nulls.NewString(*inPtr)
}

func GetCurrentUserFromGqlContext(ctx context.Context, testUser User) User {
	if testUser.ID > 0 {
		return testUser
	}

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

// FlattenPopErrors - pop validation errors are complex structures, this flattens them to a simple string
func FlattenPopErrors(popErrs *validate.Errors) string {
	var msg string
	for key, val := range popErrs.Errors {
		msg += fmt.Sprintf("%s: %s |", key, strings.Join(val, ", "))
	}

	return msg
}

// IsSqlNoRowsErr Checks if given error is a no results/rows error and therefore not really an error at all
func IsSqlNoRowsErr(err error) bool {
	if err != nil && errors.Cause(err) == sql.ErrNoRows {
		return true
	}

	return false
}

// NullsStringIsURL is a model field validator
// which makes sure that a NullsString that is not null is
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

	newV := validators.URLIsPresent{Name: v.Name, Field: value, Message: v.Message}
	newV.IsValid(errors)
}
