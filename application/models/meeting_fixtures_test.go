package models

import (
	"github.com/silinternational/wecarry-api/domain"
	"testing"
	"time"
)

func CreateMeetingFixtures(ms *ModelSuite, t *testing.T, users Users) Meetings {
	if err := ms.DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	locations := []Location{{}, {}}
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	meetings := Meetings{
		{
			CreatedByID: users[0].ID,
			Name:        "Meeting 1",
			LocationID:  locations[0].ID,
			StartDate:   time.Now(),
			EndDate:     time.Now(),
		},
		{
			CreatedByID: users[0].ID,
			Name:        "Meeting 2",
			LocationID:  locations[1].ID,
			StartDate:   time.Now(),
			EndDate:     time.Now(),
		},
	}
	for i := range meetings {
		meetings[i].UUID = domain.GetUUID()
		if err := ms.DB.Create(&meetings[i]); err != nil {
			t.Errorf("could not create test meeting ... %v", err)
			t.FailNow()
		}
		if err := ms.DB.Load(&meetings[i], "CreatedBy", "Location"); err != nil {
			t.Errorf("Error loading meeting associations: %s", err)
			t.FailNow()
		}
	}
	return meetings
}
