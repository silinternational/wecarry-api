package actions

import (
	"fmt"
	"strconv"

	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/gqlgen"
	"github.com/silinternational/wecarry-api/models"
)

type meetingQueryFixtures struct {
	models.Locations
	models.Meetings
	models.Users
	models.File
	models.Posts
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
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	CreatedBy   struct {
		Nickname string `json:"nickname"`
	} `json:"createdBy"`
	ImageFile struct {
		ID string `json:"id"`
	} `json:"imageFile"`
	Location struct {
		Country string `json:"country"`
	} `json:"location"`
	Posts []struct {
		ID string `json:"id"`
	} `json:"posts"`
}

type meetingInvitationsResponse struct {
	MeetingInvitations []meetingInvitation `json:"meetingInvitations"`
}

type meetingInvitationResponse struct {
	MeetingInvitation meetingInvitation `json:"meetingInvitation"`
}

type meetingInvitation struct {
	Meeting struct {
		ID string `json:"id"`
	} `json:"meeting"`
	Inviter struct {
		ID string `json:"id"`
	} `json:"inviter"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarURL"`
}

const allMeetingInvitationFields = "meeting {id} inviter {id} email avatarURL"

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
			startDate
			endDate
			createdBy {nickname}
			imageFile {id}
			location {country}
			posts {id}
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

	as.Equal(2, len(gotMtg.Posts), "incorrect number of meeting posts")
	as.Equal(f.Posts[1].UUID.String(), gotMtg.Posts[0].ID, "wrong post returned in meeting posts")
	as.Equal(f.Posts[0].UUID.String(), gotMtg.Posts[1].ID, "wrong post returned in meeting posts")
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
			visibility: ` + gqlgen.MeetingVisibilityInviteOnly.String() + `
		`
	query := `mutation { meeting: createMeeting(input: {` + input + `})
			{ createdBy { nickname } imageFile { id } name
			description location { description country latitude longitude }
			startDate endDate moreInfoURL }}`

	var resp meetingResponse
	as.NoError(as.testGqlQuery(query, user.Nickname, &resp))

	gotMtg := resp.Meeting

	as.True(uuid.UUID{}.String() != gotMtg.ID, "don't want empty UUID")
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

func (as *ActionSuite) Test_UpdateMeeting() {
	f := createFixturesForMeetings(as)

	var resp meetingResponse

	input := `id: "` + f.Meetings[0].UUID.String() + `" imageFileID: "` + f.File.UUID.String() + `"` +
		`
			name: "new name"
			description: "new description"
			location: {description:"new location" country:"dc" latitude:1.1 longitude:2.2}
			startDate: "2025-09-19"
			endDate: "2025-09-29"
			moreInfoURL: "new.example.com"
			visibility: ` + gqlgen.MeetingVisibilityInviteOnly.String() + `
		`
	query := `mutation { meeting: updateMeeting(input: {` + input + `}) { id imageFile { id }
			createdBy { nickname } name description
			location { description country latitude longitude}
			startDate endDate moreInfoURL }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &resp))

	err := as.DB.Load(&(f.Meetings[0]), "ImageFile")
	as.NoError(err, "failed to load meeting fixture, %s")

	gotMtg := resp.Meeting

	as.Equal(f.Meetings[0].UUID.String(), gotMtg.ID)
	as.Equal("new name", gotMtg.Name, "incorrect meeting.Name")
	as.Equal("new description", gotMtg.Description, "incorrect meeting.Description")
	as.Equal("new.example.com", gotMtg.MoreInfoURL, "incorrect meeting MoreInfoURL")
	as.Equal(f.Users[0].Nickname, gotMtg.CreatedBy.Nickname, "incorrect meeting CreatedBy")
	as.Equal("2025-09-19", gotMtg.StartDate, "incorrect meeting StartDate")
	as.Equal("2025-09-29", gotMtg.EndDate, "incorrect meeting EndDate")

	as.Equal(f.File.UUID.String(), gotMtg.ImageFile.ID)
	as.Equal("dc", gotMtg.Location.Country, "incorrect meeting Location.Country")

	// Not authorized
	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.Error(err, "expected an authorization error but did not get one")

	as.Contains(err.Error(), "You are not allowed to edit the information for that meeting.", "incorrect authorization error message")

}

func (as *ActionSuite) Test_CreateMeetingInvitations() {
	f := createFixturesForMeetings(as)

	var resp meetingInvitationsResponse

	const queryTemplate = `mutation { meetingInvitations : createMeetingInvitations(input: %s) { %s } }`

	type testCase struct {
		Name        string
		Emails      []string
		MeetingID   string
		TestUser    models.User
		GoodEmails  int
		ExpectError string
	}

	testCases := []testCase{
		{
			Name:        "empty list",
			Emails:      []string{},
			MeetingID:   f.Meetings[0].UUID.String(),
			TestUser:    f.Users[0],
			GoodEmails:  0,
			ExpectError: "",
		},
		{
			Name:        "one good, one bad",
			Emails:      []string{"email0@example.com", "email1example.com"},
			MeetingID:   f.Meetings[0].UUID.String(),
			TestUser:    f.Users[0],
			GoodEmails:  1,
			ExpectError: "email1example.com",
		},
		{
			Name:        "all good",
			Emails:      []string{"email0@example.com", "email1@example.com"},
			MeetingID:   f.Meetings[0].UUID.String(),
			TestUser:    f.Users[0],
			GoodEmails:  2,
			ExpectError: "",
		},
	}

	for _, tc := range testCases {
		emails := ""
		for _, e := range tc.Emails {
			emails = emails + `"` + e + `" `
		}
		input := fmt.Sprintf(`{ meetingID: "%s" emails: [%+v] sendEmail: false }`, tc.MeetingID, emails)

		query := fmt.Sprintf(queryTemplate, input, allMeetingInvitationFields)
		err := as.testGqlQuery(query, tc.TestUser.Nickname, &resp)

		if tc.ExpectError != "" {
			as.Error(err)
			as.Contains(err.Error(), tc.ExpectError, "didn't get expected error message")
		} else {
			as.NoError(err)
		}
		as.Equal(tc.GoodEmails, len(resp.MeetingInvitations))
		for i := range resp.MeetingInvitations {
			as.Equal(resp.MeetingInvitations[i].Email, "email"+strconv.Itoa(i)+"@example.com")
			as.Equal(resp.MeetingInvitations[i].Meeting.ID, f.Meetings[0].UUID.String())
			as.Equal(resp.MeetingInvitations[i].Inviter.ID, f.Users[0].UUID.String())
		}
	}
}
