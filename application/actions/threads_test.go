package actions

import (
	"fmt"
	"testing"
	"time"

	"github.com/silinternational/wecarry-api/api"
)

func (as *ActionSuite) TestConversations() {
	f := createFixturesForThreadQuery(as)

	tests := []struct {
		name         string
		token        string
		wantContains []string
		notWant      string
	}{
		{
			name:         "empty list",
			token:        f.Users[2].Nickname,
			wantContains: []string{"[]"},
		},
		{
			name:  "typical",
			token: f.Users[0].Nickname,
			wantContains: []string{
				fmt.Sprintf(`"id":"%s"`, f.Threads[0].UUID),
				fmt.Sprintf(`"nickname":"%s"`, f.Users[0].Nickname),
				fmt.Sprintf(`"nickname":"%s"`, f.Users[1].Nickname),
				fmt.Sprintf(`"content":"Reply from %s"`, f.Users[0].Nickname),
				fmt.Sprintf(`"content":"Message from %s"`, f.Users[1].Nickname),
				fmt.Sprintf(`"sender":{"id":"%s"`, f.Users[0].UUID),
				fmt.Sprintf(`"sender":{"id":"%s"`, f.Users[1].UUID),
				fmt.Sprintf(`"request":{"id":"%s"`, f.Requests[0].UUID),
				`"unread_message_count":1`,
			},
			notWant: `"participants":[{"id":"00000000-`,
		},
	}

	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			req := as.JSON("/threads")
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tt.token)
			req.Headers["content-type"] = "application/json"
			res := req.Get()

			body := res.Body.String()
			as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

			as.verifyResponseData(tt.wantContains, body, "In TestConversations")

			if tt.notWant != "" {
				as.NotContains(body, tt.notWant)
			}
		})
	}
}

func (as *ActionSuite) TestMarkMessagesAsRead() {
	f := createFixturesForThreadQuery(as)

	users0 := f.Users[0]
	testTime := time.Now().Add(1)

	reqBody := api.MarkMessagesAsReadInput{
		Time: testTime,
	}

	req := as.JSON("/threads/%s/read", f.Threads[0].UUID.String())
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", users0.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Put(reqBody)

	body := res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

	wantContains := []string{
		fmt.Sprintf(`"id":"%s"`, f.Threads[0].UUID),
		fmt.Sprintf(`"nickname":"%s"`, users0.Nickname),
		fmt.Sprintf(`"last_viewed_at":"%s`, testTime.Format("2006-01-02T15:04:05")),
		fmt.Sprintf(`"content":"Reply from %s"`, users0.Nickname),
		fmt.Sprintf(`"sender":{"id":"%s"`, users0.UUID),
		fmt.Sprintf(`"request":{"id":"%s"`, f.Requests[0].UUID),
	}
	as.verifyResponseData(wantContains, body, "In TestMarkMessagesAsRead")
}
