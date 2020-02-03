package actions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/silinternational/wecarry-api/domain"
)

func (as *ActionSuite) Test_serviceHandler() {
	tests := []struct {
		name        string
		token       string
		requestBody string
		wantCode    int
		wantKey     string
		wantTask    ServiceTaskName
	}{
		{
			name:     "incorrect bearer token",
			token:    "bad token",
			wantCode: http.StatusUnauthorized,
			wantKey:  domain.ErrorNotAuthenticated,
		},
		{
			name:        "malformed body",
			token:       "abc123",
			requestBody: "malformed",
			wantCode:    http.StatusBadRequest,
			wantKey:     domain.ErrorBadRequest,
		},
		{
			name:        "bad task name",
			token:       "abc123",
			requestBody: `{"task":"bad_task"}`,
			wantCode:    http.StatusUnprocessableEntity,
			wantKey:     domain.ErrorUnprocessableEntity,
		},
		{
			name:        "file cleanup",
			token:       "abc123",
			requestBody: `{"task":"file_cleanup"}`,
			wantTask:    ServiceTaskFileCleanup,
		},
		{
			name:        "token cleanup",
			token:       "abc123",
			requestBody: `{"task":"token_cleanup"}`,
			wantTask:    ServiceTaskTokenCleanup,
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

			var gqlResponse domain.AppError
			err := json.Unmarshal(responseBody, &gqlResponse)
			as.NoError(err, "response body parsing error")

			want := domain.AppError{
				Code: tt.wantCode,
				Key:  tt.wantKey,
			}

			as.Equal(want, gqlResponse, "incorrect http error response code/key")
		})
	}
}
