package models

import (
	"testing"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
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

func createFixture(t *testing.T, f interface{}) {
	err := models.DB.Create(f)
	if err != nil {
		t.Errorf("error creating %T fixture, %s", f, err)
		t.FailNow()
	}
}
