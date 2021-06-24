package actions

import (
	"fmt"
	"time"

	"github.com/silinternational/wecarry-api/api"
)

func (as *ActionSuite) TestConversations() {
	f := createFixturesForThreadQuery(as)

	users0 := f.Users[0]
	users1 := f.Users[1]

	req := as.JSON("/conversations")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", users0.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Get()

	body := res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

	wantContains := []string{
		fmt.Sprintf(`"id":"%s"`, f.Threads[0].UUID),
		fmt.Sprintf(`"nickname":"%s"`, users0.Nickname),
		fmt.Sprintf(`"nickname":"%s"`, users1.Nickname),
		fmt.Sprintf(`"content":"Reply from %s"`, users0.Nickname),
		fmt.Sprintf(`"content":"Message from %s"`, users1.Nickname),
		fmt.Sprintf(`"sender":{"id":"%s"`, users0.UUID),
		fmt.Sprintf(`"sender":{"id":"%s"`, users1.UUID),
		fmt.Sprintf(`"request":{"id":"%s"`, f.Requests[0].UUID),
		`"unread_message_count":1`,
	}
	for _, w := range wantContains {
		as.Contains(body, w)
	}
}

func (as *ActionSuite) TestMarkMessagesAsRead() {
	f := createFixturesForThreadQuery(as)

	users0 := f.Users[0]
	testTime := time.Now().Add(1)

	reqBody := api.MarkMessagesAsReadInput{
		ThreadID: f.Threads[0].UUID.String(),
		Time:     testTime,
	}

	req := as.JSON("/markMessagesAsRead")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", users0.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Post(reqBody)

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
	for _, w := range wantContains {
		as.Contains(body, w)
	}

}
