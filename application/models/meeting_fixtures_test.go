package models

import (
	"testing"
	"time"

	"github.com/silinternational/wecarry-api/domain"
)

// createMeetingFixtures_FindByUUID creates two meetings associated with the first user passed in.
func createMeetingFixtures_FindByUUID(ms *ModelSuite, t *testing.T, users Users) Meetings {
	if err := ms.DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	locations := []Location{{}, {}}
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	meetings := Meetings{
		{
			UUID:        domain.GetUUID(),
			CreatedByID: users[0].ID,
			Name:        "Meeting 1",
			LocationID:  locations[0].ID,
			StartDate:   time.Now(),
			EndDate:     time.Now(),
		},
		{
			UUID:        domain.GetUUID(),
			CreatedByID: users[0].ID,
			Name:        "Meeting 2",
			LocationID:  locations[1].ID,
			StartDate:   time.Now(),
			EndDate:     time.Now(),
		},
	}
	createFixture(ms, &meetings)
	return meetings
}

// createMeetingFixtures_FindByTime creates three meetings associated with the same user.
//  The first meeting will have dates in the past,
//  the second will have StartDate in the past and EndDate in the future,
//  the third will have dates in the future.
func createMeetingFixtures_FindByTime(ms *ModelSuite) Meetings {
	uf := createUserFixtures(ms.DB, 1)
	users := uf.Users

	locations := make(Locations, 4)
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	meetings := Meetings{
		{
			CreatedByID: users[0].ID,
			Name:        "Mtg Past",
			LocationID:  locations[0].ID,

			StartDate: time.Now().Add(-domain.DurationWeek * 10),
			EndDate:   time.Now().Add(-domain.DurationWeek * 8),
		},
		{
			CreatedByID: users[0].ID,
			Name:        "Mtg Recent",
			LocationID:  locations[1].ID,

			StartDate: time.Now().Add(-domain.DurationWeek * 4),
			EndDate:   time.Now().Add(-domain.DurationWeek * 2),
		},
		{
			CreatedByID: users[0].ID,
			Name:        "Mtg Now",
			LocationID:  locations[2].ID,
			StartDate:   time.Now().Add(-domain.DurationWeek * 2),
			EndDate:     time.Now().Add(domain.DurationWeek * 2),
		},
		{
			CreatedByID: users[0].ID,
			Name:        "Mtg Future",
			LocationID:  locations[3].ID,
			StartDate:   time.Now().Add(domain.DurationWeek * 2),
			EndDate:     time.Now().Add(domain.DurationWeek * 4),
		},
	}

	for i := range meetings {
		meetings[i].UUID = domain.GetUUID()
		createFixture(ms, &meetings[i])
	}
	return meetings
}

func createMeetingFixtures_CanUpdate(ms *ModelSuite) meetingFixtures {
	uf := createUserFixtures(ms.DB, 5)
	locations := createLocationFixtures(ms.DB, 1)

	mtgUser := uf.Users[0]

	superUser := &uf.Users[1]
	superUser.AdminRole = UserAdminRoleSuperAdmin
	ms.NoError(superUser.Save(Ctx()))

	salesUser := &uf.Users[2]
	salesUser.AdminRole = UserAdminRoleSalesAdmin
	ms.NoError(salesUser.Save(Ctx()))

	adminUser := &uf.Users[3]
	adminUser.AdminRole = UserAdminRoleAdmin
	ms.NoError(adminUser.Save(Ctx()))

	meeting := Meeting{
		CreatedByID: mtgUser.ID,
		Name:        "Mtg Past",
		LocationID:  locations[0].ID,
		StartDate:   time.Now().Add(domain.DurationWeek * 2),
		EndDate:     time.Now().Add(domain.DurationWeek * 4),
	}

	createFixture(ms, &meeting)

	return meetingFixtures{
		Meetings: Meetings{meeting},
		Users:    uf.Users,
	}
}
