package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gobuffalo/httptest"
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
