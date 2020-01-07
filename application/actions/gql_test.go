package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gobuffalo/httptest"

	"github.com/silinternational/wecarry-api/domain"
)

type gqlErrorResponse struct {
	Errors []struct {
		Message string   `json:"message"`
		Path    []string `json:"path"`
	} `json:"errors"`
	Data interface{} `json:"data"`
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

	var gqlResponse struct {
		Errors []struct {
			Message string   `json:"message"`
			Path    []string `json:"path"`
		} `json:"errors"`
		Data json.RawMessage `json:"data"`
	}
	err = json.Unmarshal(responseBody, &gqlResponse)
	as.NoError(err)

	if len(gqlResponse.Errors) > 0 {
		return fmt.Errorf("gql error: %v", gqlResponse.Errors)
	}

	return json.Unmarshal(gqlResponse.Data, &response)
}

func jsonEscapeString(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

func gqlMeetingResp(as *ActionSuite, accessToken string, httpMethod string) domain.AppError {

	query := `{ meetings {id name}}`

	body := strings.NewReader(fmt.Sprintf(`{"query":"%s"}`, jsonEscapeString(query)))
	req := httptest.NewRequest(httpMethod, "/gql", body)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()
	as.App.ServeHTTP(rr, req)

	responseBody, err := ioutil.ReadAll(rr.Body)
	as.NoError(err)

	var gqlResponse domain.AppError
	err = json.Unmarshal(responseBody, &gqlResponse)
	as.NoError(err, "unmarshalling gql error response")
	return gqlResponse
}

func (as *ActionSuite) Test_GqlBadGetQuery() {

	f := createFixturesForMeetings(as)
	user := f.Users[0]
	accessToken := user.Nickname

	httpMethod := "GET"
	gqlResponse := gqlMeetingResp(as, accessToken, httpMethod)
	want := domain.AppError{
		Code: http.StatusMethodNotAllowed,
		Key:  domain.ErrorMethodNotAllowed,
	}

	as.Equal(want, gqlResponse, "incorrect app error")
}

func (as *ActionSuite) Test_GqlBadLogin() {

	_ = createFixturesForMeetings(as)

	accessToken := "bad one"
	httpMethod := "POST"
	gqlResponse := gqlMeetingResp(as, accessToken, httpMethod)

	want := domain.AppError{
		Code: http.StatusUnauthorized,
		Key:  domain.ErrorNotAuthenticated,
	}

	as.Equal(want, gqlResponse, "incorrect app error")
}
