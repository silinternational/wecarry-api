package actions

import (
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type watchQueryFixtures struct {
	models.Users
	models.Locations
	models.Watches
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
	Location struct {
		Country string `json:"country"`
	} `json:"location"`
}

func createFixturesForWatches(as *ActionSuite) watchQueryFixtures {
	uf := test.CreateUserFixtures(as.DB, 2)
	user := uf.Users[0]
	locations := test.CreateLocationFixtures(as.DB, 2)

	watches := models.Watches{
		{
			OwnerID:    user.ID,
			LocationID: nulls.NewInt(locations[0].ID),
		},
		{
			OwnerID:    user.ID,
			LocationID: nulls.NewInt(locations[1].ID),
		},
	}

	for i := range watches {
		watches[i].UUID = domain.GetUUID()
		createFixture(as, &watches[i])
	}

	return watchQueryFixtures{
		Users:     uf.Users,
		Locations: locations,
		Watches:   watches,
	}
}

func (as *ActionSuite) Test_MyWatches() {
	f := createFixturesForWatches(as)
	watches := f.Watches

	query := `{ watches: myWatches { id owner { nickname } location { country } }}`

	var resp watchesResponse

	user := f.Users[0]

	err := as.testGqlQuery(query, user.Nickname, &resp)
	as.NoError(err)

	got := resp.Watches

	as.Equal(2, len(got), "incorrect number of Watches")
	as.Equal(watches[1].UUID.String(), got[0].ID, "incorrect Watch UUID")
	as.Equal(user.Nickname, got[1].Owner.Nickname, "incorrect Watch Owner")
	as.Equal(f.Locations[0].Country, got[1].Location.Country, "incorrect Watch Location")
}

func (as *ActionSuite) Test_CreateWatch() {
	f := createFixturesForWatches(as)
	user := f.Users[0]

	input := `location: {description:"watch location" country:"dc" latitude:1.1 longitude:2.2}`

	query := `mutation { watch: createWatch(input: {` + input + `})
		{ id owner { nickname } location { country } }}`

	var resp watchResponse
	as.NoError(as.testGqlQuery(query, user.Nickname, &resp))

	got := resp.Watch

	as.True(uuid.UUID{}.String() != got.ID, "don't want empty UUID")
	as.Equal(user.Nickname, got.Owner.Nickname, "incorrect Watch Owner")
	as.Equal("dc", got.Location.Country, "incorrect watch Location.Country")
}

func (as *ActionSuite) Test_UpdateWatch() {
	f := createFixturesForWatches(as)

	var resp watchResponse

	input := `id: "` + f.Watches[0].UUID.String() + `" ` +
		`location: {description:"new location" country:"dc" latitude:1.1 longitude:2.2}`

	query := `mutation { watch: updateWatch(input: {` + input + `})
		{ id owner { nickname } location { country } }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &resp))

	got := resp.Watch

	as.Equal(f.Watches[0].UUID.String(), got.ID)
	as.Equal(f.Users[0].Nickname, got.Owner.Nickname, "incorrect Watch Owner")
	as.Equal("dc", got.Location.Country, "incorrect Watch Location.Country")

	// Not authorized
	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.Error(err, "expected an authorization error but did not get one")

	as.Contains(err.Error(), "Watch not found",
		"incorrect authorization error message")
}

func (as *ActionSuite) Test_RemoveWatch() {
	f := createFixturesForWatches(as)

	var resp watchesResponse

	query1 := `mutation { watches: removeWatch (input: {id: "` + f.Watches[0].UUID.String() +
		`"}) { id owner { nickname } location { country } }}`

	// Not authorized
	err := as.testGqlQuery(query1, f.Users[1].Nickname, &resp)
	as.Error(err, "expected an authorization error but did not get one")
	as.Contains(err.Error(), "problem finding the Watch", "incorrect authorization error message")

	// Delete one, leave one
	as.NoError(as.testGqlQuery(query1, f.Users[0].Nickname, &resp))
	as.Equal(1, len(resp.Watches))
	as.Equal(f.Watches[1].UUID.String(), resp.Watches[0].ID)
	as.Equal(f.Users[0].Nickname, resp.Watches[0].Owner.Nickname, "incorrect Watch Owner")
	as.Equal(f.Locations[1].Country, resp.Watches[0].Location.Country, "incorrect Watch Location")

	query2 := `mutation { watches: removeWatch (input: {id: "` + f.Watches[1].UUID.String() +
		`"}) { id owner { nickname } location { country } }}`

	// Remove last Watch
	as.NoError(as.testGqlQuery(query2, f.Users[0].Nickname, &resp))
	as.Equal(0, len(resp.Watches))
}
