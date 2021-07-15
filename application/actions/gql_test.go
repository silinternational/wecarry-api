package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/silinternational/wecarry-api/api"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type gqlError struct {
	Message string            `json:"message"`
	Path    []json.RawMessage `json:"path"` // Includes strings and ints
}

type gqlErrorResponse struct {
	Errors []gqlError      `json:"errors"`
	Data   json.RawMessage `json:"data"`
}

type humanizedError struct {
	Message string
	Path    []string
}

func humanizeGQLErrors(gqlErrors []gqlError) []humanizedError {
	outErrors := []humanizedError{}
	for _, e := range gqlErrors {
		paths := []string{}
		for _, p := range e.Path {
			nextP := string(p)
			paths = append(paths, nextP)
		}
		outErrors = append(outErrors, humanizedError{
			Message: e.Message,
			Path:    paths,
		})
	}

	return outErrors
}

func (as *ActionSuite) testGqlQuery(gqlQuery, accessToken string, response interface{}) error {
	body := strings.NewReader(fmt.Sprintf(`{"query":"%s"}`, jsonEscapeString(gqlQuery)))
	req := httptest.NewRequest("POST", "/gql", body)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()
	as.App.ServeHTTP(rr, req)

	responseBody, err := ioutil.ReadAll(rr.Body)
	as.NoError(err)

	domain.Logger.Println("response: " + string(responseBody))

	var gqlResponse gqlErrorResponse
	err = json.Unmarshal(responseBody, &gqlResponse)

	as.NoError(err, "unable to unmarshall gql response")

	as.NoError(json.Unmarshal(gqlResponse.Data, &response))

	if len(gqlResponse.Errors) > 0 {
		outErrors := humanizeGQLErrors(gqlResponse.Errors)
		return fmt.Errorf("gql error: %v", outErrors)
	}

	return nil
}

func jsonEscapeString(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

func makeCall(as *ActionSuite, httpMethod, route, accessToken string, body io.Reader) []byte {
	req := httptest.NewRequest(httpMethod, route, body)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()
	as.App.ServeHTTP(rr, req)

	responseBody, err := ioutil.ReadAll(rr.Body)
	as.NoError(err)
	return responseBody
}

func gqlMeetingResp(as *ActionSuite, accessToken string, httpMethod string) api.AppError {
	query := `{ meetings {id name}}`

	body := strings.NewReader(fmt.Sprintf(`{"query":"%s"}`, jsonEscapeString(query)))

	responseBody := makeCall(as, httpMethod, "/gql", accessToken, body)

	var gqlResponse api.AppError
	err := json.Unmarshal(responseBody, &gqlResponse)
	as.NoError(err, "unmarshalling gql error response")
	return gqlResponse
}

func (as *ActionSuite) Test_GqlBadQueries() {
	t := as.T()

	f := createFixturesForMeetings(as)
	user := f.Users[0]

	tests := []struct {
		name        string
		httpMethod  string
		accessToken string
		want        api.AppError
	}{
		{
			name:        "bad because of GET",
			httpMethod:  "GET",
			accessToken: user.Nickname,
			want: api.AppError{
				Code: http.StatusMethodNotAllowed,
				Key:  api.ErrorMethodNotAllowed,
			},
		},
		{
			name:        "bad because of no auth",
			httpMethod:  "POST",
			accessToken: "Bad one",
			want: api.AppError{
				Code: http.StatusUnauthorized,
				Key:  api.ErrorNotAuthenticated,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := gqlMeetingResp(as, test.accessToken, test.httpMethod)
			as.Equal(test.want, got, "incorrect app error")
		})
	}
}

func (as *ActionSuite) Test_BadRoute() {
	f := createFixturesForMeetings(as)
	user := f.Users[0]

	accessToken := user.Nickname
	httpMethod := "POST"

	body := strings.NewReader("anything=goes")
	responseBody := makeCall(as, httpMethod, "/wonderland", accessToken, body)

	var gqlResponse api.AppError
	err := json.Unmarshal(responseBody, &gqlResponse)
	as.NoError(err, "unmarshalling gql error response")

	want := api.AppError{
		Code: http.StatusNotFound,
		Key:  api.ErrorRouteNotFound,
	}

	as.Equal(want, gqlResponse, "incorrect app error")
}

func (as *ActionSuite) locationInput(l models.Location) string {
	var geo string
	if l.Latitude.Valid && l.Longitude.Valid {
		geo = fmt.Sprintf(",latitude:%f,longitude:%f", l.Latitude.Float64, l.Longitude.Float64)
	}
	return fmt.Sprintf(`{description:"%s",country:"%s"%s}`, l.Description, l.Country, geo)
}
