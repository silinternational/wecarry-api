package gqlgen

import (
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/silinternational/wecarry-api/models"
	"github.com/stretchr/testify/suite"
)

// GqlgenSuite establishes a test suite for gqlgen tests
type GqlgenSuite struct {
	suite.Suite
}

// Test_GqlgenSuite runs the GqlgenSuite test suite
func Test_GqlgenSuite(t *testing.T) {
	suite.Run(t, new(GqlgenSuite))
}

func getGqlClient() *client.Client {
	h := handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{}}))
	srv := httptest.NewServer(h)
	c := client.New(srv.URL)
	return c
}

func createFixture(t *testing.T, f interface{}) {
	err := models.DB.Create(f)
	if err != nil {
		t.Errorf("error creating %T fixture, %s", f, err)
		t.FailNow()
	}
}
