package actions

import (
	"context"
	"io/ioutil"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/silinternational/wecarry-api/dataloader"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/gqlgen"
)

func gqlHandler(c buffalo.Context) error {
	gqlSuccess := true

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
		gqlSuccess = false
		return err
	})

	server.ServeHTTP(c.Response(), c.Request().WithContext(newCtx))

	if !gqlSuccess {
		domain.ErrLogger.Printf("error, rolling back tx in GQLHandler")
		return nil
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		domain.ErrLogger.Printf("no transaction found in GQLHandler")
	}
	if err := tx.TX.Commit(); err != nil {
		domain.ErrLogger.Printf("database commit failed, %s", err)
	}

	return nil
}
