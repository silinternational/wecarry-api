package actions

import (
	"github.com/99designs/gqlgen/handler"
	"github.com/gobuffalo/buffalo"
	"github.com/silinternational/handcarry-api/gqlgen"
)

func GQLHandler(c buffalo.Context) error {
	h := handler.GraphQL(gqlgen.NewExecutableSchema(gqlgen.Config{Resolvers: &gqlgen.Resolver{}}))
	h.ServeHTTP(c.Response(), c.Request())


	return nil
}

