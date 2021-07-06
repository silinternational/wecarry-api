package actions

import (
	"fmt"

	"github.com/silinternational/wecarry-api/api"
)

type messageResponse struct {
	Message struct {
		ID      string `json:"id"`
		Content string `json:"content"`
		Sender  struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		} `json:"sender"`
		Thread struct {
			ID           string `json:"id"`
			Participants []struct {
				ID       string `json:"id"`
				Nickname string `json:"nickname"`
			} `json:"participants"`
		} `json:"thread"`
	} `json:"message"`
}

func (as *ActionSuite) TestMessageQuery() {
	f := createFixtures_MessageQuery(as)

	query := `{ message(id: "` + f.Messages[0].UUID.String() + `")
		{ id content sender { nickname } thread {id participants {nickname}}}}`

	var resp messageResponse

	err := as.testGqlQuery(query, f.Users[0].Nickname, &resp)
	as.NoError(err)

	thread, err := f.Messages[0].GetThread(as.DB)
	as.NoError(err)

	err = thread.LoadParticipants(as.DB)
	as.NoError(err)
	as.Equal(2, len(thread.Participants), "incorrect number of thread participants")

	as.Equal(f.Messages[0].UUID.String(), resp.Message.ID)
	as.Equal(f.Messages[0].Content, resp.Message.Content)
	as.Equal(f.Users[1].Nickname, resp.Message.Sender.Nickname)
	as.Equal(thread.UUID.String(), resp.Message.Thread.ID)
	as.Equal(thread.Participants[0].Nickname, resp.Message.Thread.Participants[0].Nickname)
	as.Equal(thread.Participants[1].Nickname, resp.Message.Thread.Participants[1].Nickname)
}

func (as *ActionSuite) TestCreateMessage() {
	f := createFixtures_MessageQuery(as)

	newContent := "New Message Created"

	input := `requestID: "` + f.Requests[0].UUID.String() + `" ` +
		`threadID: "` + f.Threads[0].UUID.String() + `" content: "` + newContent + `" `
	query := `mutation { message: createMessage (input: {` + input +
		`}) { id content thread {id}}}`

	var resp messageResponse

	err := as.testGqlQuery(query, f.Users[0].Nickname, &resp)
	as.NoError(err)

	thread, err := f.Messages[0].GetThread(as.DB)
	as.NoError(err)

	err = thread.LoadMessages(as.DB)
	as.NoError(err)
	as.Equal(3, len(thread.Messages), "incorrect number of thread messages")

	as.Equal(newContent, resp.Message.Content)
	as.Equal(thread.UUID.String(), resp.Message.Thread.ID)
}

func (as *ActionSuite) TestMessagesCreate() {
	f := createFixtures_MessageQuery(as)
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
	for _, w := range wantContains {
		as.Contains(body, w)
	}
}
