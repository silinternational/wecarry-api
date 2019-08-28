package models

import (
	"context"
	"log"

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
