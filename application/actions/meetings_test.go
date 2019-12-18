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

type recentMeetingsResponse struct {
	Meetings []meeting `json:"recentmeetings"`
}

type meetingResponse struct {
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
	ImageFile struct {
		ID string `json:"id"`
	} `json:"imageFile"`
	Location struct {
		Country string `json:"country"`
	} `json:"location"`
}

func (as *ActionSuite) Test_MeetingQuery() {
	f := createFixturesForMeetings(as)
	meetings := f.Meetings

	testMtg := meetings[2]

	query := `{ meeting(id: "` + testMtg.UUID.String() + `")
		{
			id
		    name
            description
			moreInfoURL
			createdBy { nickname}
			startDate
			endDate
			imageFile {id}
			location {country}
		}}`

	var resp meetingResponse

	user := f.Users[0]

	err := as.testGqlQuery(query, user.Nickname, &resp)
	as.NoError(err)

	testLocation := f.Locations[2]
	gotMtg := resp.Meeting

	as.Equal(testMtg.UUID.String(), gotMtg.ID, "incorrect meeting UUID")
	as.Equal(testMtg.Name, gotMtg.Name, "incorrect meeting Name")
	as.Equal(testMtg.Description.String, gotMtg.Description, "incorrect meeting Description")
	as.Equal(testMtg.MoreInfoURL.String, gotMtg.MoreInfoURL, "incorrect meeting MoreInfoURL")
	as.Equal(user.Nickname, gotMtg.CreatedBy.Nickname, "incorrect meeting CreatedBy")
	as.Equal(testMtg.StartDate.Format(domain.DateFormat), gotMtg.StartDate.Format(domain.DateFormat),
		"incorrect meeting StartDate")
	as.Equal(testMtg.EndDate.Format(domain.DateFormat), gotMtg.EndDate.Format(domain.DateFormat),
		"incorrect meeting EndDate")

	image, err := testMtg.GetImage()
	as.NoError(err, "unexpected error getting ImageFile")
	wantUUID := ""
	if image != nil {
		wantUUID = image.UUID.String()
	}
	as.Equal(wantUUID, gotMtg.ImageFile.ID, "incorrect ImageFile")

	as.Equal(testLocation.Country, gotMtg.Location.Country, "incorrect meeting Location")

}

func (as *ActionSuite) Test_MeetingsQuery() {
	f := createFixturesForMeetings(as)
	meetings := f.Meetings

	query := `{ meetings
		{
			id
		    name
            description
			moreInfoURL
			createdBy { nickname}
			startDate
			endDate
			imageFile {id}
			location {country}
		}}`

	var resp meetingsResponse

	user := f.Users[0]

	err := as.testGqlQuery(query, user.Nickname, &resp)
	as.NoError(err)

	for i := 0; i < 2; i++ {
		wantMtg := meetings[i+2]
		testLocation := f.Locations[i+2]
		gotMtg := resp.Meetings[i]

		as.Equal(wantMtg.UUID.String(), gotMtg.ID, "incorrect meeting UUID")
		as.Equal(wantMtg.Name, gotMtg.Name, "incorrect meeting Name")
		as.Equal(wantMtg.Description.String, gotMtg.Description, "incorrect meeting Description")
		as.Equal(wantMtg.MoreInfoURL.String, gotMtg.MoreInfoURL, "incorrect meeting MoreInfoURL")
		as.Equal(user.Nickname, gotMtg.CreatedBy.Nickname, "incorrect meeting CreatedBy")
		as.Equal(wantMtg.StartDate.Format(domain.DateFormat), gotMtg.StartDate.Format(domain.DateFormat),
			"incorrect meeting StartDate")
		as.Equal(wantMtg.EndDate.Format(domain.DateFormat), gotMtg.EndDate.Format(domain.DateFormat),
			"incorrect meeting EndDate")

		image, err := wantMtg.GetImage()
		as.NoError(err, "unexpected error getting ImageFile")
		wantUUID := ""
		if image != nil {
			wantUUID = image.UUID.String()
		}
		as.Equal(wantUUID, gotMtg.ImageFile.ID, "incorrect ImageFile")

		as.Equal(testLocation.Country, gotMtg.Location.Country, "incorrect meeting Location")
	}
}

func (as *ActionSuite) Test_RecentMeetingsQuery() {
	f := createFixturesForMeetings(as)
	meetings := f.Meetings

	query := `{ recentMeetings
		{
			id
		    name
			startDate
			endDate
		}}`

	var resp recentMeetingsResponse

	user := f.Users[0]

	err := as.testGqlQuery(query, user.Nickname, &resp)
	as.NoError(err, "unexpected error getting recent meetings")

	as.Equal(1, len(resp.Meetings), "incorrect number of recent meetings")

	wantMtg := meetings[1]
	gotMtg := resp.Meetings[0]

	as.Equal(wantMtg.UUID.String(), gotMtg.ID, "incorrect meeting UUID")
	as.Equal(wantMtg.Name, gotMtg.Name, "incorrect meeting Name")
	as.Equal(wantMtg.StartDate.Format(domain.DateFormat), gotMtg.StartDate.Format(domain.DateFormat),
		"incorrect meeting StartDate")
	as.Equal(wantMtg.EndDate.Format(domain.DateFormat), gotMtg.EndDate.Format(domain.DateFormat),
		"incorrect meeting EndDate")
}
