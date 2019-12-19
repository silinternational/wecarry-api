package actions

import (
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type meetingQueryFixtures struct {
	models.Locations
	models.Meetings
	models.Users
	models.File
}

type meetingsResponse struct {
	Meetings []meeting `json:"meetings"`
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
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
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
	as.Equal(testMtg.StartDate.Format(domain.DateFormat), gotMtg.StartDate,
		"incorrect meeting StartDate")
	as.Equal(testMtg.EndDate.Format(domain.DateFormat), gotMtg.EndDate,
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
		as.Equal(wantMtg.StartDate.Format(domain.DateFormat), gotMtg.StartDate,
			"incorrect meeting StartDate")
		as.Equal(wantMtg.EndDate.Format(domain.DateFormat), gotMtg.EndDate,
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

	query := `{ meetings: recentMeetings
		{
		    id
		    name
		    startDate
		    endDate
		}}`

	var resp meetingsResponse

	user := f.Users[0]

	err := as.testGqlQuery(query, user.Nickname, &resp)
	as.NoError(err, "unexpected error getting recent meetings")

	as.Equal(1, len(resp.Meetings), "incorrect number of recent meetings")

	wantMtg := meetings[1]
	gotMtg := resp.Meetings[0]

	as.Equal(wantMtg.UUID.String(), gotMtg.ID, "incorrect meeting UUID")
	as.Equal(wantMtg.Name, gotMtg.Name, "incorrect meeting Name")
	as.Equal(wantMtg.StartDate.Format(domain.DateFormat), gotMtg.StartDate,
		"incorrect meeting StartDate")
	as.Equal(wantMtg.EndDate.Format(domain.DateFormat), gotMtg.EndDate,
		"incorrect meeting EndDate")
}

func (as *ActionSuite) Test_CreateMeeting() {
	f := createFixturesForMeetings(as)
	user := f.Users[0]

	input := `imageFileID: "` + f.File.UUID.String() + `"` +
		`
			name: "name"
			description: "new description"
			location: {description:"meeting location" country:"dc" latitude:1.1 longitude:2.2}
			startDate: "2025-03-01"
			endDate: "2025-03-21"
			moreInfoURL: "example.com"
		`
	query := `mutation { meeting: createMeeting(input: {` + input + `}) 
			{ createdBy { nickname } imageFile { id } name
			description location { description country latitude longitude }
			startDate endDate moreInfoURL }}`

	var resp meetingResponse
	as.NoError(as.testGqlQuery(query, user.Nickname, &resp))

	gotMtg := resp.Meeting

	emptyUUID := uuid.UUID{}
	as.NotEqual(emptyUUID, gotMtg.ID, "don't want empty UUID")
	as.Equal("name", gotMtg.Name, "incorrect meeting Name")
	as.Equal("new description", gotMtg.Description, "incorrect meeting Description")
	as.Equal("example.com", gotMtg.MoreInfoURL, "incorrect meeting MoreInfoURL")
	as.Equal(user.Nickname, gotMtg.CreatedBy.Nickname, "incorrect meeting CreatedBy")
	as.Equal("2025-03-01", gotMtg.StartDate,
		"incorrect meeting StartDate")
	as.Equal("2025-03-21", gotMtg.EndDate,
		"incorrect meeting EndDate")

	as.Equal(f.File.UUID.String(), gotMtg.ImageFile.ID, "incorrect ImageFileID")

	as.Equal("dc", gotMtg.Location.Country, "incorrect meeting Location.Country")
}
