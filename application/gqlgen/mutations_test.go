package gqlgen

import (
	"github.com/gobuffalo/suite"
)

type ActionSuite struct {
	*suite.Action
}

func (as *ActionSuite) TestFailure() {
	as.T().Error("failure test")
	as.T().FailNow()
}
