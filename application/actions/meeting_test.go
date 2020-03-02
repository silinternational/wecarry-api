package actions

import (
	"fmt"
	"strconv"
	"testing"

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
	models.MeetingInvites
	models.MeetingParticipants
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
	Invites      []meetingInvite      `json:"invites"`
	Participants []meetingParticipant `json:"participants"`
}

const allMeetingFields = `id name description moreInfoURL startDate endDate createdBy {nickname} imageFile {id}
	location {country} posts {id} invites {meeting{id} email} participants {user{id} meeting{id}}`

type meetingInvitesResponse struct {
	MeetingInvites []meetingInvite `json:"MeetingInvites"`
}

type meetingInvite struct {
	Meeting struct {
		ID string `json:"id"`
	} `json:"meeting"`
	Inviter struct {
		ID string `json:"id"`
	} `json:"inviter"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarURL"`
}

const allMeetingInviteFields = "meeting {id} inviter {id} email avatarURL"

type meetingParticipantsResponse struct {
	MeetingParticipants []meetingParticipant `json:"MeetingParticipants"`
}

type meetingParticipant struct {
	Meeting struct {
		ID string `json:"id"`
	} `json:"meeting"`
	User struct {
		ID string `json:"id"`
	} `json:"user"`
	IsOrganizer bool `json:"isOrganizer"`
	Invite      struct {
		ID string `json:"id"`
	} `json:"invite"`
}

const allMeetingParticipantFields = "meeting {id} user {id} isOrganizer invite {email}"

func (as *ActionSuite) Test_MeetingQuery() {
	f := createFixturesForMeetings(as)
	meetings := f.Meetings

	testMtg := meetings[2]

	query := `{ meeting(id: "` + testMtg.UUID.String() + `") {` + allMeetingFields + "}}"

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

	image, err := testMtg.ImageFile()
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

	as.Equal(2, len(gotMtg.Invites), "incorrect number of invites")
	for i := range gotMtg.Invites {
		as.Equal(testMtg.UUID.String(), gotMtg.Invites[i].Meeting.ID, "wrong meeting ID on invite")
		as.Equal(f.MeetingInvites[i].Email, gotMtg.Invites[i].Email, "wrong email on invite")
	}

	as.Equal(3, len(gotMtg.Participants), "wrong number of participants")
	for i := range gotMtg.Participants {
		as.Equal(f.Users[i].UUID.String(), gotMtg.Participants[i].User.ID, "wrong user ID on participant")
		as.Equal(testMtg.UUID.String(), gotMtg.Participants[i].Meeting.ID, "wrong meeting ID on participant")
	}
}

func (as *ActionSuite) Test_MeetingsQuery() {
	f := createFixturesForMeetings(as)
	meetings := f.Meetings

	query := "{ meetings {" + allMeetingFields + "}}"

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

		image, err := wantMtg.ImageFile()
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
	query := `mutation { meeting: createMeeting(input: {` + input + `}) {` + allMeetingFields + "}}"

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
	query := `mutation { meeting: updateMeeting(input: {` + input + `}) {` + allMeetingFields + "}}"

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &resp))

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
	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.Error(err, "expected an authorization error but did not get one")

	as.Contains(err.Error(), "You are not allowed to edit the information for that meeting.", "incorrect authorization error message")

}

func (as *ActionSuite) Test_CreateMeetingInvites() {
	f := createFixturesForMeetings(as)

	var resp meetingInvitesResponse

	const queryTemplate = `mutation { meetingInvites : createMeetingInvites(input: %s) { %s } }`

	type testCase struct {
		name       string
		emails     []string
		meetingID  string
		testUser   models.User
		goodEmails int
		wantErr    string
	}

	testCases := []testCase{
		{
			name:       "empty list",
			emails:     []string{},
			meetingID:  f.Meetings[0].UUID.String(),
			testUser:   f.Users[0],
			goodEmails: 0,
			wantErr:    "",
		},
		{
			name:       "one good, one bad",
			emails:     []string{"email0@example.com", "email1example.com"},
			meetingID:  f.Meetings[0].UUID.String(),
			testUser:   f.Users[0],
			goodEmails: 1,
			wantErr:    "email1example.com",
		},
		{
			name:       "all good",
			emails:     []string{"email0@example.com", "email1@example.com"},
			meetingID:  f.Meetings[0].UUID.String(),
			testUser:   f.Users[0],
			goodEmails: 2,
			wantErr:    "",
		},
		{
			name:      "not allowed",
			emails:    []string{"email0@example.com", "email1@example.com"},
			meetingID: f.Meetings[1].UUID.String(),
			testUser:  f.Users[1],
			wantErr:   "not allowed",
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			emails := ""
			for _, e := range tc.emails {
				emails = emails + `"` + e + `" `
			}
			input := fmt.Sprintf(`{ meetingID: "%s" emails: [%+v] sendEmail: false }`, tc.meetingID, emails)

			query := fmt.Sprintf(queryTemplate, input, allMeetingInviteFields)
			err := as.testGqlQuery(query, tc.testUser.Nickname, &resp)

			if tc.wantErr != "" {
				as.Error(err)
				as.Contains(err.Error(), tc.wantErr, "didn't get expected error message")
				return
			}
			as.NoError(err)
			as.Equal(tc.goodEmails, len(resp.MeetingInvites))
			for i := range resp.MeetingInvites {
				as.Equal(resp.MeetingInvites[i].Email, "email"+strconv.Itoa(i)+"@example.com")
				as.Equal(resp.MeetingInvites[i].Meeting.ID, f.Meetings[0].UUID.String())
				as.Equal(resp.MeetingInvites[i].Inviter.ID, f.Users[0].UUID.String())
			}
		})
	}
}

func (as *ActionSuite) Test_RemoveMeetingInvite() {
	f := createFixturesForMeetings(as)

	var resp meetingInvitesResponse

	const queryTemplate = `mutation { meetingInvites : removeMeetingInvite(input: %s) { %s } }`

	type testCase struct {
		name           string
		email          string
		meetingID      string
		testUser       models.User
		responseEmails []string
		wantErr        string
	}

	testCases := []testCase{
		{
			name:           "bad email",
			email:          "not_invited@example.com",
			meetingID:      f.Meetings[0].UUID.String(),
			testUser:       f.Users[0],
			responseEmails: []string{},
			wantErr:        "problem removing the meeting invite",
		},
		{
			name:           "not creator, not organizer",
			email:          "invitee2@example.com",
			meetingID:      f.Meetings[1].UUID.String(),
			testUser:       f.Users[1],
			responseEmails: []string{},
			wantErr:        "not allowed",
		},
		{
			name:           "creator",
			email:          "invitee0@example.com",
			meetingID:      f.Meetings[2].UUID.String(),
			testUser:       f.Users[0],
			responseEmails: []string{"invitee1@example.com"},
		},
		{
			name:           "organizer",
			email:          "invitee1@example.com",
			meetingID:      f.Meetings[2].UUID.String(),
			testUser:       f.Users[1],
			responseEmails: []string{},
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			input := fmt.Sprintf(`{ meetingID: "%s" email: "%s" }`, tc.meetingID, tc.email)

			query := fmt.Sprintf(queryTemplate, input, allMeetingInviteFields)
			err := as.testGqlQuery(query, tc.testUser.Nickname, &resp)

			if tc.wantErr != "" {
				as.Error(err)
				as.Contains(err.Error(), tc.wantErr, "didn't get expected error message")
				return
			}
			as.NoError(err)
			as.Equal(len(tc.responseEmails), len(resp.MeetingInvites))
			for i := range resp.MeetingInvites {
				as.Equal(tc.responseEmails[i], resp.MeetingInvites[i].Email)
				as.Equal(tc.meetingID, resp.MeetingInvites[i].Meeting.ID)
				as.Equal(f.Users[0].UUID.String(), resp.MeetingInvites[i].Inviter.ID)
			}
		})
	}
}

func (as *ActionSuite) Test_RemoveMeetingParticipant() {
	f := createFixturesForMeetings(as)

	var resp meetingParticipantsResponse

	const queryTemplate = `mutation { meetingParticipants : removeMeetingParticipant(input: %s) { %s } }`

	type testCase struct {
		name        string
		userID      string
		meetingID   string
		testUser    models.User
		responseIDs []string
		wantErr     string
	}

	testCases := []testCase{
		{
			name:        "userID not a participant",
			userID:      f.Users[0].UUID.String(),
			meetingID:   f.Meetings[1].UUID.String(),
			testUser:    f.Users[0],
			responseIDs: []string{},
			wantErr:     "problem removing the participant",
		},
		{
			name:        "test user not creator, not organizer",
			userID:      f.Users[1].UUID.String(),
			meetingID:   f.Meetings[1].UUID.String(),
			testUser:    f.Users[1],
			responseIDs: []string{},
			wantErr:     "not allowed",
		},
		{
			name:        "test user is organizer, removing participant",
			userID:      f.Users[2].UUID.String(),
			meetingID:   f.Meetings[2].UUID.String(),
			testUser:    f.Users[1],
			responseIDs: []string{f.Users[0].UUID.String(), f.Users[1].UUID.String()},
		},
		{
			name:        "test user is creator, removing organizer",
			userID:      f.Users[1].UUID.String(),
			meetingID:   f.Meetings[2].UUID.String(),
			testUser:    f.Users[0],
			responseIDs: []string{f.Users[0].UUID.String()},
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			input := fmt.Sprintf(`{ meetingID: "%s" userID: "%s" }`, tc.meetingID, tc.userID)

			query := fmt.Sprintf(queryTemplate, input, allMeetingParticipantFields)
			err := as.testGqlQuery(query, tc.testUser.Nickname, &resp)

			if tc.wantErr != "" {
				as.Error(err)
				as.Contains(err.Error(), tc.wantErr, "didn't get expected error message")
				return
			}
			as.NoError(err)
			as.Equal(len(tc.responseIDs), len(resp.MeetingParticipants))
			for i := range resp.MeetingParticipants {
				as.Equal(tc.responseIDs[i], resp.MeetingParticipants[i].User.ID)
				as.Equal(tc.meetingID, resp.MeetingParticipants[i].Meeting.ID)
			}
		})
	}
}
