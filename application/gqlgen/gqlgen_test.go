package gqlgen

import (
	"testing"

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
