package actions

import (
	"fmt"

	"github.com/silinternational/wecarry-api/api"
)

func (as *ActionSuite) TestMessagesCreate() {
	f := createFixturesForMessagesCreate(as)
	user0 := f.Users[0]
	message0 := f.Messages[0]
	thread0 := f.Threads[0]
	request0 := f.Requests[0]

	newContent := "New Message Created"
	threadID := thread0.UUID.String()
	reqBody := api.MessageInput{
		Content:   newContent,
		RequestID: request0.UUID.String(),
		ThreadID:  &threadID,
	}

	req := as.JSON("/messages")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", user0.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Post(reqBody)

	body := res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

	thread, err := message0.GetThread(as.DB)
	as.NoError(err)

	err = thread.LoadMessages(as.DB)
	as.NoError(err)
	as.Equal(3, len(thread.Messages), "incorrect number of thread messages")

	wantContains := []string{
		fmt.Sprintf(`"id":"%s"`, thread0.UUID),
		fmt.Sprintf(`"nickname":"%s"`, user0.Nickname),
		fmt.Sprintf(`"content":"Reply from %s"`, user0.Nickname),
		fmt.Sprintf(`"content":"%s"`, newContent),
		fmt.Sprintf(`"sender":{"id":"%s"`, user0.UUID),
		fmt.Sprintf(`"request":{"id":"%s"`, request0.UUID),
		fmt.Sprintf(`"title":"%s","description":"%s"`, request0.Title, request0.Description.String),
	}
	as.verifyResponseData(wantContains, body, "In TestMessagesCreate")
}
