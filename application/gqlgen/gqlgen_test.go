package gqlgen

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GqlgenSuite struct {
	suite.Suite
}

func Test_ActionSuite(t *testing.T) {
	suite.Run(t, new(GqlgenSuite))
}
