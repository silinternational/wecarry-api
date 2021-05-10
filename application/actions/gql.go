package actions

import (
	"context"
	"io/ioutil"

	"github.com/99designs/gqlgen/handler"
	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/dataloader"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/gqlgen"
)

func gqlHandler(c buffalo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err == nil {
		c.Logger().Debugf("request body: %s", body)
	}

	h := handler.GraphQL(gqlgen.NewExecutableSchema(gqlgen.Config{Resolvers: &gqlgen.Resolver{}}))
	newCtx := context.WithValue(c.Request().Context(), domain.BuffaloContext, c)
	newCtx = dataloader.GetDataLoaderContext(newCtx)

	h.ServeHTTP(c.Response(), c.Request().WithContext(newCtx))

	return nil
}
