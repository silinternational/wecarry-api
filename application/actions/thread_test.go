package actions

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
		Post struct {
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
	as.Equal(f.Posts[0].UUID.String(), resp.Threads[0].Post.ID)
	as.Equal(f.Messages[0].UUID.String(), resp.Threads[0].Messages[0].ID)
	as.Equal(f.Messages[0].Content, resp.Threads[0].Messages[0].Content)
	as.Equal(f.Users[1].Nickname, resp.Threads[0].Messages[0].Sender.Nickname)

	thread := f.Threads[0]
	err = thread.Load("Participants")
	as.NoError(err)

	participants, err := thread.GetParticipants()
	as.NoError(err)
	as.Equal(2, len(participants), "incorrect number of thread participants")

	as.Equal(participants[0].Nickname, resp.Threads[0].Participants[0].Nickname)
	as.Equal(participants[1].Nickname, resp.Threads[0].Participants[1].Nickname)
}
