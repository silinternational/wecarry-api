package actions

import (
	"context"

	"github.com/99designs/gqlgen/handler"
	"github.com/gobuffalo/buffalo"
	"github.com/silinternational/handcarry-api/gqlgen"
)

func GQLHandler(c buffalo.Context) error {
	h := handler.GraphQL(gqlgen.NewExecutableSchema(gqlgen.Config{Resolvers: &gqlgen.Resolver{}}))
	newCtx := context.WithValue(c.Request().Context(), "BuffaloContext", c)
	h.ServeHTTP(c.Response(), c.Request().WithContext(newCtx))

	return nil
}
