package actions

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) verifyMeeting(expected models.Meeting, actual api.Meeting, msg string) {
	as.Equal(expected.UUID, actual.ID, msg+", ID is not correct")
	as.Equal(expected.Name, actual.Name, msg+", Name is not correct")
	as.Equal(expected.Description.String, actual.Description, msg+", Description is not correct")
	as.Equal(expected.StartDate.Format(domain.DateFormat), actual.StartDate, msg+", StartDate is not correct")
	as.Equal(expected.EndDate.Format(domain.DateFormat), actual.EndDate, msg+", EndDate is not correct")
	as.True(expected.CreatedAt.Equal(actual.CreatedAt), msg+", CreatedAt is not correct")
	as.True(expected.UpdatedAt.Equal(actual.UpdatedAt), msg+", UpdatedAt is not correct")
	as.Equal(expected.MoreInfoURL.String, actual.MoreInfoURL, msg+", MoreInfoURL is not correct")
}

func (as *ActionSuite) Test_meetingsList() {
	f := createFixturesForMeetings(as)
	mtgs := f.Meetings
	lctns := f.Locations
	user := f.Users[0]

	req := as.JSON("/events")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", user.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Get()

	body := res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

	wantContains := []string{
		fmt.Sprintf(`"nickname":"%s"`, user.Nickname),
		fmt.Sprintf(`"participants":[]`),
		`"url":"http://minio:9000/wca-test-bucket`,
	}

	for i := 2; i < 4; i++ {
		lctn := lctns[i]
		moreContains := []string{
			fmt.Sprintf(`"id":"%s"`, mtgs[i].UUID.String()),
			fmt.Sprintf(`"is_editable":%t`, mtgs[i].CanUpdate(user)),
			fmt.Sprintf(`"name":"%s"`, mtgs[i].Name),
			fmt.Sprintf(`"start_date":"%s`, mtgs[i].StartDate.Format(domain.DateFormat)),
			fmt.Sprintf(`"end_date":"%s`, mtgs[i].EndDate.Format(domain.DateFormat)),
			fmt.Sprintf(`"location":{"description":"%s"`, lctn.Description),
			fmt.Sprintf(`"country":"%s"`, lctn.Country),
			fmt.Sprintf(`"latitude":%s`, convertFloat64ToIntString(lctn.Latitude)),
			fmt.Sprintf(`"longitude":%s`, convertFloat64ToIntString(lctn.Longitude)),
		}
		wantContains = append(wantContains, moreContains...)
	}

	as.verifyResponseData(wantContains, body, "In Test_MeetingsList")

	as.NotContains(body, mtgs[0].Name, "should not have included name of past meeting")
	as.NotContains(body, mtgs[1].Name, "should not have included name of recent meeting")
}

func (as *ActionSuite) Test_meetingsCreate() {
	f := createFixturesForMeetings(as)

	nextWeek := time.Now().UTC().Add(domain.DurationWeek)
	weekAfterNext := time.Now().UTC().Add(domain.DurationWeek * 2)

	goodMeeting := api.MeetingInput{
		Name:        "Good Meeting",
		Description: nulls.NewString("This is a good meeting"),
		StartDate:   nextWeek.Format(domain.DateFormat),
		EndDate:     weekAfterNext.Format(domain.DateFormat),
		MoreInfoURL: nulls.NewString("http://events.example.org/1"),
		Location:    locationX,
		ImageFileID: nulls.NewUUID(f.File.UUID),
	}
	badMeetingLocation := api.MeetingInput{
		StartDate: "2030-01-02",
		EndDate:   "2030-01-12",
	}

	badMeetingFile := goodMeeting
	badMeetingFile.ImageFileID = nulls.NewUUID(domain.GetUUID())

	badMeetingStartDate := goodMeeting
	badMeetingStartDate.StartDate = "2020-01-02 12:01:02"

	tests := []struct {
		name            string
		user            models.User
		meeting         api.MeetingInput
		wantStatus      int
		wantErrContains string
	}{
		{
			name:            "authn error",
			user:            models.User{},
			meeting:         goodMeeting,
			wantStatus:      http.StatusUnauthorized,
			wantErrContains: api.ErrorNotAuthenticated.String(),
		},
		{
			name:            "bad meeting location input",
			user:            f.Users[1],
			meeting:         badMeetingLocation,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: api.ErrorLocationCreateFailure.String(),
		},
		{
			name:            "bad file input",
			user:            f.Users[1],
			meeting:         badMeetingFile,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: api.ErrorMeetingImageIDNotFound.String(),
		},
		{
			name:            "bad start date input",
			user:            f.Users[1],
			meeting:         badMeetingStartDate,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: api.ErrorMeetingInvalidStartDate.String(),
		},
		{
			name:       "good input",
			user:       f.Users[1],
			meeting:    goodMeeting,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			req := as.JSON("/events")
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tt.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Post(&tt.meeting)

			body := res.Body.String()
			as.Equal(tt.wantStatus, res.Code, "incorrect status code returned, body: %s", body)

			if tt.wantStatus != http.StatusOK {
				if tt.wantErrContains != "" {
					as.Contains(body, tt.wantErrContains, "missing error message")
				}
				return
			}

			wantData := []string{
				`"name":"` + tt.meeting.Name,
				`"created_by":{"id":"` + tt.user.UUID.String(),
				`"location":{"description":"` + locationX.Description,
				`"start_date":"` + nextWeek.Format(domain.DateFormat),
				`"end_date":"` + weekAfterNext.Format(domain.DateFormat),
				`"more_info_url":"` + tt.meeting.MoreInfoURL.String,
				`"image_file":{"id":"` + tt.meeting.ImageFileID.UUID.String(),
			}
			as.verifyResponseData(wantData, body, "")
		})
	}
}

func (as *ActionSuite) Test_meetingsUpdate() {
	f := createFixturesForMeetings(as)

	nextMonth := time.Now().UTC().Add(domain.DurationWeek * 4)
	monthAfterNext := time.Now().UTC().Add(domain.DurationWeek * 8)

	goodMeeting := api.MeetingInput{
		Name:        "Good Meeting New",
		Description: nulls.NewString("This is a new good meeting"),
		StartDate:   nextMonth.Format(domain.DateFormat),
		EndDate:     monthAfterNext.Format(domain.DateFormat),
		MoreInfoURL: nulls.NewString("http://events.example.org/1/new"),
		Location:    locationX,
		ImageFileID: nulls.NewUUID(f.File.UUID),
	}

	badMeetingLocation := api.MeetingInput{
		Name:      "badMeetingLocation",
		StartDate: "2030-01-02",
		EndDate:   "2030-01-12",
	}

	badMeetingFile := goodMeeting
	badMeetingFile.ImageFileID = nulls.NewUUID(domain.GetUUID())

	badMeetingStartDate := goodMeeting
	badMeetingStartDate.StartDate = "2020-01-02 12:01:02"

	tests := []struct {
		name            string
		user            models.User
		input           api.MeetingInput
		meeting         models.Meeting
		wantStatus      int
		wantErrContains string
	}{
		{
			name:            "authn error",
			user:            models.User{},
			input:           goodMeeting,
			meeting:         f.Meetings[0],
			wantStatus:      http.StatusUnauthorized,
			wantErrContains: api.ErrorNotAuthenticated.String(),
		},
		{
			name:            "bad input location",
			user:            f.Users[1],
			input:           badMeetingLocation,
			meeting:         f.Meetings[0],
			wantStatus:      http.StatusBadRequest,
			wantErrContains: api.ErrorLocationCreateFailure.String(),
		},
		{
			name:            "bad input file",
			user:            f.Users[1],
			input:           badMeetingFile,
			meeting:         f.Meetings[0],
			wantStatus:      http.StatusBadRequest,
			wantErrContains: api.ErrorMeetingImageIDNotFound.String(),
		},
		{
			name:            "authz error",
			user:            f.Users[1],
			input:           goodMeeting,
			meeting:         f.Meetings[0],
			wantStatus:      http.StatusNotFound,
			wantErrContains: api.ErrorNotAuthorized.String(),
		},
		{
			name:       "good input",
			user:       f.Users[0],
			input:      goodMeeting,
			meeting:    f.Meetings[0],
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			req := as.JSON("/events/" + tt.meeting.UUID.String())
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tt.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Put(&tt.input)

			body := res.Body.String()
			as.Equal(tt.wantStatus, res.Code, "incorrect status code returned, body: %s", body)

			if tt.wantStatus != http.StatusOK {
				if tt.wantErrContains != "" {
					as.Contains(body, tt.wantErrContains, "missing error message")
				}
				return
			}

			wantData := []string{
				`"name":"` + tt.input.Name,
				`"created_by":{"id":"` + tt.user.UUID.String(),
				`"location":{"description":"` + locationX.Description,
				`"start_date":"` + nextMonth.Format(domain.DateFormat),
				`"end_date":"` + monthAfterNext.Format(domain.DateFormat),
				`"more_info_url":"` + tt.input.MoreInfoURL.String,
				`"image_file":{"id":"` + tt.input.ImageFileID.UUID.String(),
			}
			as.verifyResponseData(wantData, body, "")
		})
	}
}

func (as *ActionSuite) Test_meetingsJoin() {
	f := createFixturesForMeetings(as)

	mtgCreator := f.Users[0]

	type testCase struct {
		name            string
		inviteCode      string
		location        models.Location
		meeting         models.Meeting
		user            models.User
		wantInvite      models.MeetingInvite
		wantHTTPStatus  int
		wantContainsErr string
	}

	testCases := []testCase{
		{
			name:            "bad Meeting ID",
			meeting:         models.Meeting{UUID: domain.GetUUID()},
			user:            f.Users[0],
			wantHTTPStatus:  http.StatusNotFound,
			wantContainsErr: api.ErrorMeetingsGet.String(),
		},
		{
			name:           "already a participant",
			location:       f.Locations[2],
			meeting:        f.Meetings[2],
			user:           f.Users[2],
			wantHTTPStatus: http.StatusOK,
		},
		{
			name:           "meeting creator",
			location:       f.Locations[0],
			meeting:        f.Meetings[0],
			user:           f.Users[0],
			wantHTTPStatus: http.StatusOK,
		},
		{
			name:           "regular user",
			inviteCode:     f.MeetingInvites[3].Secret.String(),
			location:       f.Locations[3],
			meeting:        f.Meetings[3],
			user:           f.Users[1],
			wantInvite:     f.MeetingInvites[3],
			wantHTTPStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			reqBody := api.MeetingParticipantInput{
				MeetingID: tc.meeting.UUID.String(),
			}
			if tc.inviteCode != "" {
				reqBody.Code = &tc.inviteCode
			}

			req := as.JSON("/events/join")
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tc.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Post(reqBody)

			body := res.Body.String()
			as.Equal(tc.wantHTTPStatus, res.Code, "incorrect status code returned, body: %s", body)
			if tc.wantHTTPStatus != http.StatusOK {
				as.Contains(body, tc.wantContainsErr, "missing error message")
				return
			}

			wantContains := []string{
				fmt.Sprintf(`"created_by":{"id":"%s"`, mtgCreator.UUID),
				fmt.Sprintf(`"participants":[]`),
				fmt.Sprintf(`"created_by":{"id":"%s"`, mtgCreator.UUID.String()),
				fmt.Sprintf(`"id":"%s"`, tc.meeting.UUID.String()),
				fmt.Sprintf(`"name":"%s"`, tc.meeting.Name),
				fmt.Sprintf(`"start_date":"%s`, tc.meeting.StartDate.Format(domain.DateFormat)),
				fmt.Sprintf(`"end_date":"%s`, tc.meeting.EndDate.Format(domain.DateFormat)),
				fmt.Sprintf(`"location":{"description":"%s"`, tc.location.Description),
				fmt.Sprintf(`"country":"%s"`, tc.location.Country),
				fmt.Sprintf(`"latitude":%s`, convertFloat64ToIntString(tc.location.Latitude)),
				fmt.Sprintf(`"longitude":%s`, convertFloat64ToIntString(tc.location.Longitude)),
			}

			as.verifyResponseData(wantContains, body, "In Test_meetingsJoin")
		})
	}
}

func (as *ActionSuite) Test_meetingsGet() {
	f := createFixturesForMeetings(as)

	mtgCreator := f.Users[0]
	mtgParticipant := f.Users[1]

	testCases := []struct {
		name             string
		user             models.User
		meeting          models.Meeting
		wantStatus       int
		wantParticipants bool
	}{
		{
			name:       "authn error",
			user:       models.User{},
			meeting:    f.Meetings[0],
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad meeting ID",
			user:       mtgCreator,
			meeting:    models.Meeting{UUID: uuid.UUID{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "meeting not Found",
			user:       mtgCreator,
			meeting:    models.Meeting{UUID: domain.GetUUID()},
			wantStatus: http.StatusNotFound,
		},
		{
			name:             "good for creator",
			user:             mtgCreator,
			meeting:          f.Meetings[1],
			wantStatus:       http.StatusOK,
			wantParticipants: true,
		},
		{
			name:             "good for participant but no participants",
			user:             mtgParticipant,
			meeting:          f.Meetings[1],
			wantStatus:       http.StatusOK,
			wantParticipants: false,
		},
	}
	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			req := as.JSON("/events/" + tc.meeting.UUID.String())
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tc.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Get()

			body := res.Body.String()
			as.Equal(tc.wantStatus, res.Code, "incorrect status code returned, body: %s", body)

			if tc.wantStatus != http.StatusOK {
				return
			}

			as.NoError(as.DB.Load(&tc.meeting, "Location"), "error in test trying to load meeting location")

			wantContains := []string{
				fmt.Sprintf(`"id":"%s"`, tc.meeting.UUID.String()),
				fmt.Sprintf(`"created_by":{"id":"%s"`, mtgCreator.UUID.String()),
				fmt.Sprintf(`"nickname":"%s"`, mtgCreator.Nickname),
				fmt.Sprintf(`"name":"%s"`, tc.meeting.Name),
				fmt.Sprintf(`"start_date":"%s`, tc.meeting.StartDate.Format(domain.DateFormat)),
				fmt.Sprintf(`"end_date":"%s`, tc.meeting.EndDate.Format(domain.DateFormat)),
				fmt.Sprintf(`"location":{"description":"%s"`, tc.meeting.Location.Description),
				fmt.Sprintf(`"country":"%s"`, tc.meeting.Location.Country),
				fmt.Sprintf(`"latitude":%s`, convertFloat64ToIntString(tc.meeting.Location.Latitude)),
				fmt.Sprintf(`"longitude":%s`, convertFloat64ToIntString(tc.meeting.Location.Longitude)),
			}

			as.verifyResponseData(wantContains, body, "In Test_meetingsGet")

			if tc.wantParticipants {
				wantContains := fmt.Sprintf(`"participants":[{"user":{"id":"%s"`, mtgParticipant.UUID.String())
				as.Contains(body, wantContains, "incorrect participants list")
			} else {
				wantContains := fmt.Sprintf(`"participants":[]`)
				as.Contains(body, wantContains, "participants list should be empty")
			}
		})
	}
}
