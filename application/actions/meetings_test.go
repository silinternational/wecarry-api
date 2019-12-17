package actions

import (
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"time"
)

type meetingQueryFixtures struct {
	models.Locations
	models.Meetings
	models.Users
}

type meetingsResponse struct {
	Meetings []meeting `json:"meetings"`
}

type MeetingResponse struct {
	Meeting meeting `json:"meeting"`
}

type meeting struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MoreInfoURL string `json:"moreInfoURL"`
	CreatedBy   struct {
		Nickname string `json:"nickname"`
	} `json:"createdBy"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Location  struct {
		Country string `json:"country"`
	} `json:"location"`
}

func (as *ActionSuite) Test_MeetingsQuery() {
	f := createFixturesForMeetings(as)
	meetings := f.Meetings

	//query := `{ meeting(id: "` + testMtg.UUID.String() + `")
	query := `{ meetings
		{
			id
		    name
            description
			moreInfoURL
			createdBy { nickname}
			startDate
			endDate
			location {country}
		}}`

	var resp meetingsResponse

	user := f.Users[0]

	err := as.testGqlQuery(query, user.Nickname, &resp)
	as.NoError(err)

	for i := 0; i < 2; i++ {
		testMtg := meetings[i+2]
		testLocation := f.Locations[i+2]
		gotMtg := resp.Meetings[i]

		as.Equal(testMtg.UUID.String(), gotMtg.ID, "incorrect meeting UUID")
		as.Equal(testMtg.Name, gotMtg.Name, "incorrect meeting Name")
		as.Equal(testMtg.Description.String, gotMtg.Description, "incorrect meeting Description")
		as.Equal(testMtg.MoreInfoURL.String, gotMtg.MoreInfoURL, "incorrect meeting MoreInfoURL")
		as.Equal(user.Nickname, gotMtg.CreatedBy.Nickname, "incorrect meeting CreatedBy")
		as.Equal(testMtg.StartDate.Format(domain.DateFormat), gotMtg.StartDate.Format(domain.DateFormat),
			"incorrect meeting StartDate")
		as.Equal(testMtg.EndDate.Format(domain.DateFormat), gotMtg.EndDate.Format(domain.DateFormat),
			"incorrect meeting EndDate")
		as.Equal(testLocation.Country, gotMtg.Location.Country, "incorrect meeting Location")
	}
}
