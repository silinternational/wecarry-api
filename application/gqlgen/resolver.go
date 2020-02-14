//go:generate go run github.com/99designs/gqlgen

package gqlgen

import (
	"context"
)

// Resolver is required by gqlgen
type Resolver struct{}

// Mutation is required by gqlgen
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query is required by gqlgen
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) SystemConfig(ctx context.Context) (*SystemConfig, error) {
	return nil, nil
}
