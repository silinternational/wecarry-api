//go:generate go run github.com/99designs/gqlgen

package gqlgen

import (
	"github.com/silinternational/wecarry-api/models"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// TestUser is intended as a way to inject a "current User" for unit tests
var TestUser models.User

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }
