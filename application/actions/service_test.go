package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/job"
)

func (as *ActionSuite) Test_serviceHandler() {

	postBody := func(name string) string {
		return fmt.Sprintf(`{"task":"%s"}`, name)
	}

	tests := []struct {
		name        string
		token       string
		requestBody string
		wantCode    int
		wantKey     api.ErrorKey
		wantTask    ServiceTaskName
	}{
		{
			name:     "incorrect bearer token",
			token:    "bad token",
			wantCode: http.StatusUnauthorized,
			wantKey:  api.ErrorNotAuthenticated,
		},
		{
			name:        "malformed body",
			token:       domain.Env.ServiceIntegrationToken,
			requestBody: "malformed",
			wantCode:    http.StatusBadRequest,
			wantKey:     api.ErrorBadRequest,
		},
		{
			name:        "bad task name",
			token:       domain.Env.ServiceIntegrationToken,
			requestBody: postBody("bad_task"),
			wantCode:    http.StatusUnprocessableEntity,
			wantKey:     api.ErrorUnprocessableEntity,
		},
		{
			name:        "file cleanup",
			token:       domain.Env.ServiceIntegrationToken,
			requestBody: postBody(job.FileCleanup),
			wantTask:    ServiceTaskFileCleanup,
		},
		{
			name:        "token cleanup",
			token:       domain.Env.ServiceIntegrationToken,
			requestBody: postBody(job.TokenCleanup),
			wantTask:    ServiceTaskTokenCleanup,
		},
		{
			name:        "outdated requests",
			token:       domain.Env.ServiceIntegrationToken,
			requestBody: postBody(job.OutdatedRequests),
			wantTask:    ServiceTaskOutdatedRequests,
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			var logOutput bytes.Buffer
			domain.Logger.SetOutput(&logOutput)
			var errOutput bytes.Buffer
			domain.ErrLogger.SetOutput(&errOutput)

			defer func() {
				domain.Logger.SetOutput(os.Stdout)
				domain.ErrLogger.SetOutput(os.Stderr)
			}()

			body := strings.NewReader(tt.requestBody)
			responseBody := makeCall(as, "POST", "/service", tt.token, body)

			if tt.wantCode == 0 {
				as.Contains(logOutput.String(), tt.wantTask, "didn't see expected log output")
				as.Equal("", errOutput.String(), "unexpected err log output")
				return
			}

			var appError api.AppError
			err := json.Unmarshal(responseBody, &appError)
			as.NoError(err, "response body parsing error")

			want := api.AppError{
				Code: tt.wantCode,
				Key:  tt.wantKey,
			}

			as.Equal(want, appError, "incorrect http error response code/key")
		})
	}
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
