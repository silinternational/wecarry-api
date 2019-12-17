package models

import (
	"fmt"
	"github.com/silinternational/wecarry-api/domain"
	"testing"
	"time"
)

// createMeetingFixtures creates two meetings associated with the first user passed in.
func createMeetingFixtures(ms *ModelSuite, t *testing.T, users Users) Meetings {
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
		createFixture(ms, &meetings[i])
	}
	return meetings
}

// createMeetingFixtures_TestMeetings_FindFuture creates three meetings associated with the same user.
//  The first meeting will have dates in the past,
//  the second will have StartDate in the past and EndDate in the future,
//  the third will have dates in the future.
func createMeetingFixtures_TestMeetings_FindFuture(ms *ModelSuite, t *testing.T, users Users) Meetings {
	if err := ms.DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	locations := make(Locations, 3)
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	meetings := Meetings{
		{
			CreatedByID: users[0].ID,
			Name:        "Mtg Past",
			LocationID:  locations[0].ID,
			// About 10,000 minutes per week
			StartDate: time.Now().Add(time.Duration(-time.Minute * 40000)),
			EndDate:   time.Now().Add(time.Duration(-time.Minute * 20000)),
		},
		{
			CreatedByID: users[0].ID,
			Name:        "Mtg Now",
			LocationID:  locations[1].ID,
			StartDate:   time.Now().Add(time.Duration(-time.Minute * 20000)),
			EndDate:     time.Now().Add(time.Duration(time.Minute * 20000)),
		},
		{
			CreatedByID: users[0].ID,
			Name:        "Mtg Future",
			LocationID:  locations[2].ID,
			StartDate:   time.Now().Add(time.Duration(time.Minute * 20000)),
			EndDate:     time.Now().Add(time.Duration(time.Minute * 40000)),
		},
	}

	for i := range meetings {
		fmt.Printf("  %s  %v\n", meetings[i].Name, meetings[i].EndDate)
		meetings[i].UUID = domain.GetUUID()
		createFixture(ms, &meetings[i])
	}
	return meetings
}
