package marketing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// TestSuite establishes a test suite for domain tests
type TestSuite struct {
	suite.Suite
}

// Test_TestSuite runs the test suite
func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (ts *TestSuite) Test_callApi() {
	t := ts.T()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/", apiRequestHandler)

	type args struct {
		apiRequests []ApiRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "callApi tests",
			wantErr: false,
			args: args{
				apiRequests: []ApiRequest{
					{
						Method:      http.MethodGet,
						URL:         srv.URL,
						Body:        `{"key": "value"}`,
						ContentType: "application/json",
						Username:    "user",
						Password:    "pass",
						Headers:     nil,
					},
					{
						Method:      http.MethodPost,
						URL:         srv.URL,
						Body:        `{"key": "value"}`,
						ContentType: "application/json",
						Username:    "user",
						Password:    "pass",
						Headers: map[string]string{
							"something": "else",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, ar := range tt.args.apiRequests {
				got, err := callApi(ar)
				if (err != nil) != tt.wantErr {
					t.Errorf("callApi() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				var resp ApiRequest
				err = json.Unmarshal([]byte(got), &resp)
				if err != nil {
					t.Errorf("unable to unmarshal api response: %s", err.Error())
					return
				}
				// ignore response headers for test
				ar.Headers = map[string]string{}
				resp.Headers = map[string]string{}
				if !reflect.DeepEqual(resp, ar) {
					t.Errorf("callApi() got = %v, want %v", resp, ar)
				}
			}
		})
	}
}

func apiRequestHandler(res http.ResponseWriter, req *http.Request) {
	reqBody, _ := ioutil.ReadAll(req.Body)
	un, pw, _ := req.BasicAuth()
	headers := map[string]string{}
	for key, vals := range req.Header {
		headers[key] = strings.Join(vals, " ")
	}

	requestUrl := "http://" + req.Host

	apiRequest := ApiRequest{
		Method:      req.Method,
		URL:         requestUrl,
		Body:        string(reqBody),
		ContentType: req.Header.Get("content-type"),
		Username:    un,
		Password:    pw,
		Headers:     headers,
	}

	respBody, _ := json.Marshal(apiRequest)

	res.WriteHeader(200)
	res.Header().Set("content-type", "application/json")
	_, _ = fmt.Fprintf(res, string(respBody))
}

func (ts *TestSuite) TestAddUserToList() {
	t := ts.T()

	type args struct {
		u          models.User
		apiBaseURL string
		listId     string
		username   string
		password   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test integration",
			args: args{
				u: models.User{
					Email:     "integration-test@domainasdfasdf.com",
					FirstName: "Integration",
					LastName:  "Test",
				},
				apiBaseURL: domain.Env.MailChimpAPIBaseURL,
				listId:     domain.Env.MailChimpListID,
				username:   domain.Env.MailChimpUsername,
				password:   domain.Env.MailChimpAPIKey,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddUserToList(tt.args.u, tt.args.apiBaseURL, tt.args.listId, tt.args.username, tt.args.password); (err != nil) != tt.wantErr {
				t.Errorf("AddUserToList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
