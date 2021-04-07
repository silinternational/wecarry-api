package gqlgen

import (
	"testing"

	"github.com/gobuffalo/suite/v3"
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
