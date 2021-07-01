package actions

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type watchFixtures struct {
	models.Users
	models.Locations
	models.Watches
	models.Meetings
}

func createFixturesForWatches(as *ActionSuite) watchFixtures {
	// make 2 users, 1 that has Watches, and another that will try to mess with those Watches
	uf := test.CreateUserFixtures(as.DB, 2)
	locations := test.CreateLocationFixtures(as.DB, 3)
	watches := make(models.Watches, 2)
	for i := range watches {
		watches[i].OwnerID = uf.Users[0].ID
		watches[i].DestinationID = nulls.NewInt(locations[i].ID)
		test.MustCreate(as.DB, &watches[i])
	}
	meetings := models.Meetings{
		{
			CreatedByID: uf.Users[0].ID,
			Name:        "Mtg",
			LocationID:  locations[2].ID,

			StartDate: time.Now().Add(domain.DurationWeek * 8),
			EndDate:   time.Now().Add(domain.DurationWeek * 10),
		},
	}

	for i := range meetings {
		meetings[i].UUID = domain.GetUUID()
		createFixture(as, &meetings[i])
	}

	return watchFixtures{
		Users:     uf.Users,
		Locations: locations,
		Watches:   watches,
		Meetings:  meetings,
	}
}

func (as *ActionSuite) Test_MyWatches() {
	f := createFixturesForWatches(as)
	watches := f.Watches

	owner := f.Users[0]
	destinations := f.Locations

	req := as.JSON("/watches")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", owner.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Get()

	body := res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

	wantContains := []string{
		fmt.Sprintf(`"id":"%s"`, watches[0].UUID.String()),
		fmt.Sprintf(`"id":"%s"`, watches[1].UUID.String()),
		fmt.Sprintf(`"destination":{"description":"%s"`, destinations[0].Description),
		fmt.Sprintf(`"country":"%s"`, destinations[0].Country),
		fmt.Sprintf(`"latitude":%v`, int(destinations[0].Latitude.Float64)),
		fmt.Sprintf(`"longitude":%v`, int(destinations[0].Longitude.Float64)),
		fmt.Sprintf(`"destination":{"description":"%s"`, destinations[1].Description),
		fmt.Sprintf(`"country":"%s"`, destinations[1].Country),
		fmt.Sprintf(`"latitude":%v`, int(destinations[1].Latitude.Float64)),
		fmt.Sprintf(`"longitude":%v`, int(destinations[1].Longitude.Float64)),
	}
	for _, w := range wantContains {
		as.Contains(body, w)
	}

	as.NotContains(body, `"origin":`)
	as.NotContains(body, `"meeting":`)

	// Try with no watches
	nonOwner := f.Users[1]
	req = as.JSON("/watches")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", nonOwner.Nickname)
	req.Headers["content-type"] = "application/json"
	res = req.Get()

	body = res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)
	as.Equal("[]\n", body, "expected an empty list in the response")
}

func (as *ActionSuite) Test_WatchRemove() {
	f := createFixturesForWatches(as)
	owner := f.Users[0]
	notOwner := f.Users[1]
	watches := f.Watches

	type testCase struct {
		name         string
		watchID      string
		user         models.User
		failMsg      string
		wantStatus   int
		wantContains string
	}

	testCases := []testCase{
		{
			name:         "Bad ID",
			watchID:      "badid",
			user:         owner,
			wantStatus:   http.StatusBadRequest,
			wantContains: api.ErrorMustBeAValidUUID.String(),
			failMsg:      "expected an error about a bad id",
		},
		{
			name:         "unauthorized",
			watchID:      watches[0].UUID.String(),
			user:         notOwner,
			wantStatus:   http.StatusNotFound,
			wantContains: api.ErrorNotAuthorized.String(),
			failMsg:      "expected a not authorized error",
		},
		{
			name:         "Delete one leave one",
			watchID:      watches[0].UUID.String(),
			user:         owner,
			wantStatus:   http.StatusOK,
			wantContains: watches[0].UUID.String(),
			failMsg:      "expected success with the watch uuid returned",
		},
		{
			name:         "Delete the last one",
			watchID:      watches[1].UUID.String(),
			user:         owner,
			wantStatus:   http.StatusOK,
			wantContains: watches[1].UUID.String(),
			failMsg:      "expected success with the watch uuid returned",
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			req := as.JSON("/watches/%s", tc.watchID)
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tc.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Delete()

			as.Equal(tc.wantStatus, res.Code, "incorrect response status code")

			body := res.Body.String()

			as.Contains(body, tc.wantContains, tc.failMsg)
		})
	}
}
