package models

import (
	"fmt"
	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
	"github.com/gobuffalo/pop"
	"testing"
)

func BounceTestDB() error {
	test, err := pop.Connect("test")
	if err != nil {
		return err
	}

	// drop the test db:
	err = test.Dialect.DropDB()
	if err != nil {
		return err
	}

	// create the test db:
	err = test.Dialect.CreateDB()
	if err != nil {
		return err
	}

	fm, err := pop.NewFileMigrator("../migrations", test)
	if err != nil {
		return err
	}

	if err := fm.Up(); err != nil {
		return err
	}

	return nil
}

func CreateOrgs(fixtures Organizations) error {
	for _, f := range fixtures {
		if err := DB.Create(&f); err != nil {
			return fmt.Errorf("error creating org %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func CreateUsers(fixtures Users) error {
	for _, f := range fixtures {
		if err := DB.Create(&f); err != nil {
			return fmt.Errorf("error creating user %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func CreateUserOrgs(fixtures UserOrganizations) error {
	for _, f := range fixtures {
		if err := DB.Create(&f); err != nil {
			return fmt.Errorf("error creating user-org %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func CreatePosts(fixtures Posts) error {
	for _, f := range fixtures {
		if err := DB.Create(&f); err != nil {
			return fmt.Errorf("error creating post %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func CreateThreads(fixtures Threads) error {
	db := DB
	for _, f := range fixtures {
		if err := db.Create(&f); err != nil {
			return fmt.Errorf("error creating thread %+v ...\n %v \n", f, err)
		}
	}

	threads := []Thread{}
	if err := db.All(&threads); err != nil {
		return fmt.Errorf("error retrieving new threads ... %v \n", err)
	}

	if len(threads) < len(fixtures) {
		return fmt.Errorf("wrong number of threads created, expected %v, but got %v", len(fixtures), len(threads))
	}

	return nil
}

func CreateThreadParticipants(fixtures ThreadParticipants) error {
	for _, f := range fixtures {
		if err := DB.Create(&f); err != nil {
			return fmt.Errorf("error creating threadparticipant %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func CreateMessages(fixtures Messages) error {
	for _, f := range fixtures {
		if err := DB.Create(&f); err != nil {
			return fmt.Errorf("error creating message %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func CreateUserAccessTokens(fixtures UserAccessTokens) error {
	for _, f := range fixtures {
		if err := DB.Create(&f); err != nil {
			return fmt.Errorf("error creating user access token %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

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
