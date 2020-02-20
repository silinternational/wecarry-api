package models

import (
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/suite"
)

type ModelSuite struct {
	*suite.Model
}

func Test_ModelSuite(t *testing.T) {
	model := suite.NewModel()

	as := &ModelSuite{
		Model: model,
	}
	suite.Run(t, as)
}

func createFixture(ms *ModelSuite, f interface{}) {
	err := ms.DB.Create(f)
	if err != nil {
		ms.T().Errorf("error creating %T fixture, %s", f, err)
		ms.T().FailNow()
	}
}

type testBuffaloContext struct {
	buffalo.DefaultContext
	params map[string]interface{}
}

func (b *testBuffaloContext) Value(key interface{}) interface{} {
	return b.params[key.(string)]
}

func (b *testBuffaloContext) Set(key string, val interface{}) {
	b.params[key] = val
}
