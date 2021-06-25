package actions

import (
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type watchGQLQueryFixtures struct {
	models.Users
	models.Locations
	models.Watches
	models.Meetings
}

type watchesResponse struct {
	Watches []watch `json:"watches"`
}

type watchResponse struct {
	Watch watch `json:"watch"`
}

type watch struct {
	ID    string `json:"id"`
	Owner struct {
		Nickname string `json:"nickname"`
	} `json:"owner"`
	Name        string
	Destination location `json:"destination"`
	Origin      location `json:"origin"`
	Meeting     struct {
		ID string `json:"id"`
	}
	SearchText string `json:"searchText"`
	Size       string `json:"size"`
}

type location struct {
	Country     string  `json:"country"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type watchInput struct {
	id          *string
	name        string
	destination locationInput
	origin      locationInput
	meetingID   string
	searchText  string
	size        models.RequestSize
}

type locationInput struct {
	description string
	country     string
	latitude    float64
	longitude   float64
}

const allWatchFields = `
    id
    owner { nickname }
    name
    destination { description country latitude longitude }
    origin { description country latitude longitude }
    meeting { id }
    searchText
    size
	`

func createFixturesForGQLWatches(as *ActionSuite) watchGQLQueryFixtures {
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

	return watchGQLQueryFixtures{
		Users:     uf.Users,
		Locations: locations,
		Watches:   watches,
		Meetings:  meetings,
	}
}

func (as *ActionSuite) Test_GQLMyWatches() {
	f := createFixturesForGQLWatches(as)
	watches := f.Watches

	query := "{ watches: myWatches { " + allWatchFields + "}}"

	var resp watchesResponse

	user := f.Users[0]

	err := as.testGqlQuery(query, user.Nickname, &resp)
	as.NoError(err)

	got := resp.Watches

	as.Equal(2, len(got), "incorrect number of Watches")
	as.Equal(watches[1].UUID.String(), got[0].ID, "incorrect Watch UUID")
	as.Equal(user.Nickname, got[1].Owner.Nickname, "incorrect Watch Owner")
	as.Equal(f.Locations[0].Country, got[1].Destination.Country, "incorrect Watch Destination")
}

func (as *ActionSuite) Test_CreateWatch() {
	f := createFixturesForGQLWatches(as)
	user := f.Users[0]

	type testCase struct {
		name        string
		watch       watchInput
		testUser    models.User
		expectError bool
	}

	var resp watchResponse

	testCases := []testCase{
		{
			name: "all fields",
			watch: watchInput{
				name: "foo",
				destination: locationInput{
					description: "watch destination",
					country:     "dc",
					latitude:    1.1,
					longitude:   2.2,
				},
				meetingID:  f.Meetings[0].UUID.String(),
				searchText: "search",
				size:       models.RequestSizeXlarge,
			},
			testUser: f.Users[0],
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			query := "mutation { watch: createWatch(input: {" + as.watchInputString(tc.watch) + "}) {" + allWatchFields + "}}"
			resp = watchResponse{}
			err := as.testGqlQuery(query, tc.testUser.Nickname, &resp)

			if tc.expectError {
				as.Error(err, "didn't get expected error")
				return
			}
			as.NoError(err, "unexpected error")

			as.True(uuid.UUID{}.String() != resp.Watch.ID, "don't want empty UUID")
			as.Equal(user.Nickname, resp.Watch.Owner.Nickname, "incorrect Watch Owner")
			as.Equal(tc.watch.name, resp.Watch.Name, "incorrect Watch name")
			as.Equal(tc.watch.destination.country, resp.Watch.Destination.Country, "incorrect watch Country")
			as.Equal(tc.watch.destination.description, resp.Watch.Destination.Description, "incorrect watch Description")
			as.Equal(tc.watch.destination.latitude, resp.Watch.Destination.Latitude, "incorrect watch Latitude")
			as.Equal(tc.watch.destination.longitude, resp.Watch.Destination.Longitude, "incorrect watch Longitude")
			as.Equal(tc.watch.meetingID, resp.Watch.Meeting.ID, "incorrect Watch meeting ID")
			as.Equal(tc.watch.searchText, resp.Watch.SearchText, "incorrect Watch search text")

			var dbWatch models.Watch
			err = as.DB.Where("uuid = ?", resp.Watch.ID).First(&dbWatch)
			as.NoError(err, "didn't find Watch in database")
		})
	}
}

func (as *ActionSuite) Test_UpdateWatch() {
	f := createFixturesForGQLWatches(as)
	user := f.Users[0]

	type testCase struct {
		name        string
		watch       watchInput
		testUser    models.User
		expectError string
	}

	var resp watchResponse

	watchUUID := f.Watches[0].UUID.String()

	testCases := []testCase{
		{
			name: "all fields",
			watch: watchInput{
				id:   &watchUUID,
				name: "foo",
				destination: locationInput{
					description: "new location",
					country:     "dc",
					latitude:    1.1,
					longitude:   2.2,
				},
				meetingID:  f.Meetings[0].UUID.String(),
				searchText: "search",
				size:       models.RequestSizeXlarge,
			},
			testUser: f.Users[0],
		},
		{
			name: "unauthorized",
			watch: watchInput{
				id:         &watchUUID,
				name:       "foo",
				searchText: "search",
			},
			testUser:    f.Users[1],
			expectError: "Alert not found",
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			query := "mutation { watch: updateWatch(input: {" + as.watchInputString(tc.watch) + "}) {" + allWatchFields + "}}"
			resp = watchResponse{}
			err := as.testGqlQuery(query, tc.testUser.Nickname, &resp)

			if tc.expectError != "" {
				as.Error(err, "didn't get expected error")
				as.Contains(err.Error(), tc.expectError, "incorrect error message")
				return
			}
			as.NoError(err, "unexpected error")

			as.Equal(*tc.watch.id, resp.Watch.ID, "incorrect Watch UUID")
			as.Equal(user.Nickname, resp.Watch.Owner.Nickname, "incorrect Watch Owner")
			as.Equal(tc.watch.name, resp.Watch.Name, "incorrect Watch name")
			as.Equal(tc.watch.destination.country, resp.Watch.Destination.Country, "incorrect watch Country")
			as.Equal(tc.watch.destination.description, resp.Watch.Destination.Description, "incorrect watch Description")
			as.Equal(tc.watch.destination.latitude, resp.Watch.Destination.Latitude, "incorrect watch Latitude")
			as.Equal(tc.watch.destination.longitude, resp.Watch.Destination.Longitude, "incorrect watch Longitude")
			as.Equal(tc.watch.meetingID, resp.Watch.Meeting.ID, "incorrect Watch meeting ID")
			as.Equal(tc.watch.searchText, resp.Watch.SearchText, "incorrect Watch search text")

			var dbWatch models.Watch
			err = as.DB.Where("uuid = ?", resp.Watch.ID).First(&dbWatch)
			as.NoError(err, "didn't find Watch in database")
		})
	}
}

func (as *ActionSuite) watchInputString(watch watchInput) string {
	input := ""
	if watch.id != nil {
		input = `id: "` + *watch.id + `"`
	}

	if watch.size == "" {
		watch.size = models.RequestSizeXlarge
	}

	input = fmt.Sprintf(`%s name: "%s" destination: {description:"%s" country:"%s" latitude:%f longitude:%f}
		meetingID: "%s" searchText: "%s" size: %s`,
		input, watch.name, watch.destination.description, watch.destination.country, watch.destination.latitude,
		watch.destination.longitude, watch.meetingID, watch.searchText, watch.size)

	return input
}

func (as *ActionSuite) Test_RemoveWatch() {
	f := createFixturesForGQLWatches(as)

	var resp watchesResponse

	query1 := `mutation { watches: removeWatch (input: {id: "` + f.Watches[0].UUID.String() +
		`"}) { ` + allWatchFields + "}}"

	// Not authorized
	err := as.testGqlQuery(query1, f.Users[1].Nickname, &resp)
	as.Error(err, "expected an authorization error but did not get one")
	as.Contains(err.Error(), "problem finding the Alert", "incorrect authorization error message")

	// Delete one, leave one
	as.NoError(as.testGqlQuery(query1, f.Users[0].Nickname, &resp))
	as.Equal(1, len(resp.Watches))
	as.Equal(f.Watches[1].UUID.String(), resp.Watches[0].ID)
	as.Equal(f.Users[0].Nickname, resp.Watches[0].Owner.Nickname, "incorrect Watch Owner")
	as.Equal(f.Locations[1].Country, resp.Watches[0].Destination.Country, "incorrect Watch Destination")

	query2 := `mutation { watches: removeWatch (input: {id: "` + f.Watches[1].UUID.String() +
		`"}) {` + allWatchFields + "}}"

	// Remove last Watch
	as.NoError(as.testGqlQuery(query2, f.Users[0].Nickname, &resp))
	as.Equal(0, len(resp.Watches))
}
