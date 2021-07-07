package actions

import (
	"fmt"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
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
		wantContains = append(wantContains, fmt.Sprintf(`"id":"%s"`, mtgs[i].UUID.String()))
		wantContains = append(wantContains, fmt.Sprintf(`"name":"%s"`, mtgs[i].Name))
		wantContains = append(wantContains, fmt.Sprintf(`"start_date":"%s`, mtgs[i].StartDate.Format(domain.DateFormat)))
		wantContains = append(wantContains, fmt.Sprintf(`"end_date":"%s`, mtgs[i].EndDate.Format(domain.DateFormat)))
		wantContains = append(wantContains, fmt.Sprintf(`"location":{"description":"%s"`, lctn.Description))
		wantContains = append(wantContains, fmt.Sprintf(`"country":"%s"`, lctn.Country))
		wantContains = append(wantContains, fmt.Sprintf(`"latitude":%v`, int(lctn.Latitude.Float64)))
		wantContains = append(wantContains, fmt.Sprintf(`"longitude":%v`, int(lctn.Longitude.Float64)))
	}

	for _, w := range wantContains {
		as.Contains(body, w)
	}

	as.NotContains(body, mtgs[0].Name, "should not have included name of past meeting")
	as.NotContains(body, mtgs[1].Name, "should not have included name of recent meeting")

}
