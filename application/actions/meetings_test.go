package actions

import (
	"fmt"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"net/http"
	"testing"
)

func (as *ActionSuite) verifyMeeting(expected models.Meeting, actual api.Meeting, msg string) {
	as.Equal(expected.UUID, actual.ID, msg+", ID is not correct")
	as.Equal(expected.Name, actual.Name, msg+", Name is not correct")
	as.Equal(expected.Description.String, actual.Description, msg+", Description is not correct")
	as.True(expected.StartDate.Equal(actual.StartDate), msg+", StartDate is not correct")
	as.True(expected.EndDate.Equal(actual.EndDate), msg+", EndDate is not correct")
	as.True(expected.CreatedAt.Equal(actual.CreatedAt), msg+", CreatedAt is not correct")
	as.True(expected.UpdatedAt.Equal(actual.UpdatedAt), msg+", UpdatedAt is not correct")
	as.Equal(expected.MoreInfoURL.String, actual.MoreInfoURL, msg+", MoreInfoURL is not correct")
}

func (as *ActionSuite) Test_MeetingsList() {
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
		fmt.Sprintf(`"participants":[{"user":{"id":"%s"`, user.UUID.String()),
		`"url":"http://minio:9000/wca-test-bucket`,
	}

	for i := 2; i < 4; i++ {
		lctn := lctns[i]
		moreContains := []string{
			fmt.Sprintf(`"id":"%s"`, mtgs[i].UUID.String()),
			fmt.Sprintf(`"name":"%s"`, mtgs[i].Name),
			fmt.Sprintf(`"start_date":"%s`, mtgs[i].StartDate.Format(domain.DateFormat)),
			fmt.Sprintf(`"end_date":"%s`, mtgs[i].EndDate.Format(domain.DateFormat)),
			fmt.Sprintf(`"location":{"description":"%s"`, lctn.Description),
			fmt.Sprintf(`"country":"%s"`, lctn.Country),
			fmt.Sprintf(`"latitude":%s`, convertFloat64ToIntString(lctn.Latitude.Float64)),
			fmt.Sprintf(`"longitude":%s`, convertFloat64ToIntString(lctn.Longitude.Float64)),
		}
		wantContains = append(wantContains, moreContains...)
	}

	as.verifyResponseData(wantContains, body, "In Test_MeetingsList")

	as.NotContains(body, mtgs[0].Name, "should not have included name of past meeting")
	as.NotContains(body, mtgs[1].Name, "should not have included name of recent meeting")

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

			req := as.JSON("/events")
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
				fmt.Sprintf(`"participants":[{"user":{"id":"%s"`, tc.user.UUID.String()),
				fmt.Sprintf(`"nickname":"%s"`, tc.user.Nickname),
				fmt.Sprintf(`"id":"%s"`, tc.meeting.UUID.String()),
				fmt.Sprintf(`"name":"%s"`, tc.meeting.Name),
				fmt.Sprintf(`"start_date":"%s`, tc.meeting.StartDate.Format(domain.DateFormat)),
				fmt.Sprintf(`"end_date":"%s`, tc.meeting.EndDate.Format(domain.DateFormat)),
				fmt.Sprintf(`"location":{"description":"%s"`, tc.location.Description),
				fmt.Sprintf(`"country":"%s"`, tc.location.Country),
				fmt.Sprintf(`"latitude":%s`, convertFloat64ToIntString(tc.location.Latitude.Float64)),
				fmt.Sprintf(`"longitude":%s`, convertFloat64ToIntString(tc.location.Longitude.Float64)),
			}

			as.verifyResponseData(wantContains, body, "In Test_meetingsJoin")
		})
	}
}
