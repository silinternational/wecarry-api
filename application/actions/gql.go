package actions

import (
	"context"
	"io/ioutil"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gobuffalo/buffalo"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/silinternational/wecarry-api/dataloader"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/gqlgen"
)

func gqlHandler(c buffalo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err == nil {
		c.Logger().Debugf("request body: %s", body)
	}

	server := handler.NewDefaultServer(gqlgen.NewExecutableSchema(gqlgen.Config{Resolvers: &gqlgen.Resolver{}}))
	newCtx := context.WithValue(c.Request().Context(), domain.BuffaloContext, c)
	newCtx = dataloader.GetDataLoaderContext(newCtx)

	server.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)
		domain.Error(c, "graphql error: "+err.Error())
		return err
	})

	server.ServeHTTP(c.Response(), c.Request().WithContext(newCtx))

	return nil
}
