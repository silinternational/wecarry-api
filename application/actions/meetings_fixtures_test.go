package actions

import (
	"github.com/silinternational/wecarry-api/domain"
	"time"

	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

func createFixturesForMeetings(as *ActionSuite) meetingQueryFixtures {
	uf := test.CreateUserFixtures(as.DB, 1)
	user := uf.Users[0]
	locations := test.CreateLocationFixtures(as.DB, 4)

	meetings := models.Meetings{
		{
			CreatedByID: user.ID,
			Name:        "Mtg Past",
			LocationID:  locations[0].ID,

			StartDate: time.Now().Add(time.Duration(-domain.DurationWeek * 10)),
			EndDate:   time.Now().Add(time.Duration(-domain.DurationWeek * 8)),
		},
		{
			CreatedByID: user.ID,
			Name:        "Mtg Recent",
			LocationID:  locations[1].ID,

			StartDate: time.Now().Add(time.Duration(-domain.DurationWeek * 4)),
			EndDate:   time.Now().Add(time.Duration(-domain.DurationWeek * 2)),
		},
		{
			CreatedByID: user.ID,
			Name:        "Mtg Now",
			LocationID:  locations[2].ID,
			StartDate:   time.Now().Add(time.Duration(-domain.DurationWeek * 2)),
			EndDate:     time.Now().Add(time.Duration(domain.DurationWeek * 2)),
		},
		{
			CreatedByID: user.ID,
			Name:        "Mtg Future",
			LocationID:  locations[3].ID,
			StartDate:   time.Now().Add(time.Duration(domain.DurationWeek * 2)),
			EndDate:     time.Now().Add(time.Duration(domain.DurationWeek * 4)),
		},
	}

	for i := range meetings {
		meetings[i].UUID = domain.GetUUID()
		createFixture(as, &meetings[i])
	}
	return meetingQueryFixtures{
		Locations: locations,
		Meetings:  meetings,
		Users:     uf.Users,
	}
}
