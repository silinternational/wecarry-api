package models

import (
	"fmt"
	"testing"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
	"github.com/gobuffalo/pop"
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

func CreateOrgs(fixtures []Organization) error {
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
	tablesInOrder := []string{
		"user_access_tokens",
		"user_organizations",
		"organization_domains",
		"thread_participants",
		"messages",
		"threads",
		"posts",
		"organizations",
		"users",
	}

	for _, table := range tablesInOrder {
		dq := fmt.Sprintf("delete from %s", table)
		aq := fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", table)

		err := models.DB.RawQuery(dq).Exec()
		if err != nil {
			t.Errorf("Failed to delete all %s for test, error: %s", table, err)
			t.FailNow()
		}
		err = models.DB.RawQuery(aq).Exec()
		if err != nil {
			t.Errorf("Failed to reset sequence on %s for test, error: %s", table, err)
			t.FailNow()
		}
	}
}
