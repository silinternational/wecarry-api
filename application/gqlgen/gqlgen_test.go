package gqlgen

import (
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/gobuffalo/suite"
)

// GqlgenSuite establishes a test suite for gqlgen tests
type GqlgenSuite struct {
	*suite.Model
}

// Test_GqlgenSuite runs the GqlgenSuite test suite
func Test_GqlgenSuite(t *testing.T) {
	model := suite.NewModel()

	gs := &GqlgenSuite{
		Model: model,
	}
	suite.Run(t, gs)
}

func getGqlClient() *client.Client {
	h := handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{}}))
	srv := httptest.NewServer(h)
	c := client.New(srv.URL)
	return c
}

func createFixture(gs *GqlgenSuite, f interface{}) {
	err := gs.DB.Create(f)
	if err != nil {
		gs.T().Errorf("error creating %T fixture, %s", f, err)
		gs.T().FailNow()
	}
}
