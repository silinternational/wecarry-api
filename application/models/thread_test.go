package models

import (
	"testing"

	"github.com/silinternational/handcarry-api/domain"
)

func Test_Thread(t *testing.T) {
	t.Fatal("This test needs to be implemented!")
}

func CreateThreadFixtures(t *testing.T, post Post) []Thread {
	// Load Thread test fixtures
	threads := []Thread{
		{
			Uuid:   domain.GetUuid(),
			PostID: post.ID,
		},
		{
			Uuid:   domain.GetUuid(),
			PostID: post.ID,
		},
	}
	for i := range threads {
		if err := DB.Create(&threads[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}
	return threads
}
