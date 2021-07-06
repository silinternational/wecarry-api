package actions

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"

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

type locationInput struct {
	Description string
	Country     string
	Latitude    float64
	Longitude   float64
}

type watchInput struct {
	Name        string
	Destination *locationInput
	Origin      *locationInput
	MeetingID   string
	SearchText  string
	Size        api.RequestSize
}

func (as *ActionSuite) Test_WatchesCreate() {
	f := createFixturesForWatches(as)
	owner := f.Users[0]
	notOwner := f.Users[1]

	type testCase struct {
		name         string
		watch        watchInput
		user         models.User
		wantStatus   int
		wantContains string
	}

	testCases := []testCase{
		{
			name: "bad meeting id",
			watch: watchInput{
				Name:      "Bad Meeting",
				MeetingID: domain.GetUUID().String(),
			},
			user:         notOwner,
			wantStatus:   http.StatusBadRequest,
			wantContains: api.WatchInputMeetingFailure.String(),
		},
		{
			name: "just give the name field",
			watch: watchInput{
				Name: "Empty Fields",
			},
			user:         owner,
			wantStatus:   http.StatusBadRequest,
			wantContains: api.WatchInputEmpty.String(),
		},
		{
			name: "just give the search text field",
			watch: watchInput{
				Name:       "Just Search Text",
				SearchText: "OneField",
			},
			user:       owner,
			wantStatus: http.StatusOK,
		},
		{
			name: "give all fields",
			watch: watchInput{
				Name: "AllFields",
				Destination: &locationInput{
					Description: "good watch destination",
					Country:     "dc",
					Latitude:    1.1,
					Longitude:   2.2,
				},
				Origin: &locationInput{
					Description: "good watch origin",
					Country:     "cd",
					Latitude:    11.1,
					Longitude:   22.2,
				},
				MeetingID:  f.Meetings[0].UUID.String(),
				SearchText: "AllFields",
				Size:       api.RequestSizeXlarge,
			},
			user:       owner,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			req := as.JSON("/watches")
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tc.user.Nickname)
			req.Headers["content-type"] = "application/json"

			res := req.Post(tc.watch)

			body := res.Body.String()
			as.Equal(tc.wantStatus, res.Code, "incorrect status code returned, body: %s", body)
			if tc.wantContains != "" {
				as.Contains(body, tc.wantContains)
			}

			if tc.wantStatus != http.StatusOK {
				return
			}

			as.Contains(body, `"id":"`, "results don't include the watch id")

			// Extract the uuid from the response
			parts := strings.Split(body, ":")
			as.True(len(parts) == 2, "results don't have exactly one colon")

			uuidParts := strings.Split(parts[1], `"`)
			as.True(len(uuidParts) == 3, `results don't have uuid surrounded by /"`)

			// ensure uuid is not null
			newUuid := uuidParts[1]
			as.True(uuid.UUID{}.String() != newUuid, "don't want empty UUID")

			// Get the new watch from the database and validate its values
			var dbWatch models.Watch
			err := as.DB.Eager("Destination", "Origin", "Meeting").Where("uuid = ?", newUuid).First(&dbWatch)
			as.NoError(err, "didn't find Watch in database")

			as.Equal(tc.user.ID, dbWatch.OwnerID, "incorrect Watch owner")
			as.Equal(tc.watch.Name, dbWatch.Name, "incorrect Watch name")
			if tc.watch.Destination != nil {
				as.Equal(tc.watch.Destination.Country, dbWatch.Destination.Country, "incorrect Watch Destination Country")
				as.Equal(tc.watch.Destination.Description, dbWatch.Destination.Description, "incorrect Watch Destination")
				as.Equal(tc.watch.Destination.Latitude, dbWatch.Destination.Latitude.Float64, "incorrect Watch Destination Latitude")
				as.Equal(tc.watch.Destination.Longitude, dbWatch.Destination.Longitude.Float64, "incorrect Watch Destination Longitude")
			}
			if tc.watch.Origin != nil {
				as.Equal(tc.watch.Origin.Description, dbWatch.Origin.Description, "incorrect Watch Origin")
			}

			if tc.watch.MeetingID == "" {
				as.False(dbWatch.MeetingID.Valid, "expected a null Watch MeetingID")
			} else {
				as.Equal(tc.watch.MeetingID, dbWatch.Meeting.UUID.String(), "incorrect Watch Meeting")
			}

			as.Equal(tc.watch.SearchText, dbWatch.SearchText.String, "incorrect Watch search text")
			as.Equal(tc.watch.Size.String(), dbWatch.Size.String(), "incorrect Watch search text")
		})
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
