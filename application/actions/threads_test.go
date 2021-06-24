package actions

import (
	"fmt"
	"time"

	"github.com/silinternational/wecarry-api/api"
)

type threadsResponse struct {
	Threads []struct {
		ID           string `json:"id"`
		Participants []struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		} `json:"participants"`
		Messages []struct {
			ID      string `json:"id"`
			Content string `json:"content"`
			Sender  struct {
				ID       string `json:"id"`
				Nickname string `json:"nickname"`
			} `json:"sender"`
		} `json:"messages"`
		Request struct {
			ID string `json:"id"`
		} `json:"request"`
	} `json:"threads"`
}

const allThreadFields = "id request { id } participants {nickname} messages {id content sender { nickname }}}"

func (as *ActionSuite) TestThreadsQuery() {
	f := createFixturesForThreadQuery(as)
	query := "{ threads {" + allThreadFields + "}"
	// for now, the "threads" and "myThreads" queries are the same
	testThreadsQuery(as, f, query)
}

func (as *ActionSuite) TestMyThreadsQuery() {
	f := createFixturesForThreadQuery(as)
	query := "{ threads: myThreads {" + allThreadFields + "}"
	testThreadsQuery(as, f, query)
}

func testThreadsQuery(as *ActionSuite, f threadQueryFixtures, query string) {
	var resp threadsResponse

	err := as.testGqlQuery(query, f.Users[0].Nickname, &resp)

	as.NoError(err)

	as.Equal(f.Threads[0].UUID.String(), resp.Threads[0].ID)
	as.Equal(f.Requests[0].UUID.String(), resp.Threads[0].Request.ID)
	as.Equal(f.Messages[0].UUID.String(), resp.Threads[0].Messages[0].ID)
	as.Equal(f.Messages[0].Content, resp.Threads[0].Messages[0].Content)
	as.Equal(f.Users[1].Nickname, resp.Threads[0].Messages[0].Sender.Nickname)

	thread := f.Threads[0]
	err = thread.Load(as.DB, "Participants")
	as.NoError(err)

	err = thread.LoadParticipants(as.DB)
	as.NoError(err)
	as.Equal(2, len(thread.Participants), "incorrect number of thread participants")

	as.Equal(thread.Participants[0].Nickname, resp.Threads[0].Participants[0].Nickname)
	as.Equal(thread.Participants[1].Nickname, resp.Threads[0].Participants[1].Nickname)
}

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
