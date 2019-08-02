package models

import (
	"testing"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
)

func resetTables(t *testing.T) {
	resetUsersTable(t)
	resetOrganizationsTable(t)
}

func resetUsersTable(t *testing.T) {
	// delete all existing users
	err := models.DB.RawQuery("delete from users").Exec()
	if err != nil {
		t.Errorf("Failed to delete all users for test, error: %s", err)
		t.FailNow()
	}
	err = models.DB.RawQuery("ALTER SEQUENCE users_id_seq RESTART WITH 1").Exec()
	if err != nil {
		t.Errorf("Failed to delete all users for test, error: %s", err)
		t.FailNow()
	}
}

func resetOrganizationsTable(t *testing.T) {
	// delete all existing users
	err := models.DB.RawQuery("delete from organizations").Exec()
	if err != nil {
		t.Errorf("Failed to delete all organizations for test, error: %s", err)
		t.FailNow()
	}
	err = models.DB.RawQuery("ALTER SEQUENCE organizations_id_seq RESTART WITH 1").Exec()
	if err != nil {
		t.Errorf("Failed to delete all organizations for test, error: %s", err)
		t.FailNow()
	}
}
