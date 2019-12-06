package gqlgen

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
		} `json:"post"`
	} `json:"threads"`
}

func (gs *GqlgenSuite) TestThreadsQuery() {
	f := createFixtures_ThreadQuery(gs)
	c := getGqlClient()

	query := `{ threads
		{ id post { id } participants {nickname} messages {id content sender { nickname }}}  }`

	var resp threadsResponse

	TestUser = f.Users[0]
	err := c.Post(query, &resp)

	gs.NoError(err)

	gs.Equal(f.Threads[0].UUID.String(), resp.Threads[0].ID)
	gs.Equal(f.Posts[0].UUID.String(), resp.Threads[0].Post.ID)
	gs.Equal(f.Messages[0].UUID.String(), resp.Threads[0].Messages[0].ID)
	gs.Equal(f.Messages[0].Content, resp.Threads[0].Messages[0].Content)
	gs.Equal(f.Users[1].Nickname, resp.Threads[0].Messages[0].Sender.Nickname)

	thread := f.Threads[0]
	err = thread.Load("Participants")
	gs.NoError(err)

	participants, err := thread.GetParticipants()
	gs.NoError(err)
	gs.Equal(2, len(participants), "incorrect number of thread participants")

	gs.Equal(participants[0].Nickname, resp.Threads[0].Participants[0].Nickname)
	gs.Equal(participants[1].Nickname, resp.Threads[0].Participants[1].Nickname)
}

type myThreadsResponse struct {
	MyThreads []struct {
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
				Email    string `json:"email"`
				Nickname string `json:"nickname"`
				Location struct {
					Description string `json:"description"`
				} `json:"location"`
			} `json:"sender"`
		} `json:"messages"`
	} `json:"myThreads"`
}

func (gs *GqlgenSuite) TestMyThreadsQuery() {
	f := createFixtures_ThreadQuery(gs)
	c := getGqlClient()

	query := `{ myThreads
		{ id participants {nickname} messages {id content sender { nickname }}}  }`

	var resp myThreadsResponse

	TestUser = f.Users[0]
	err := c.Post(query, &resp)

	gs.NoError(err)

	gs.Equal(f.Threads[0].UUID.String(), resp.MyThreads[0].ID)
	gs.Equal(f.Messages[0].UUID.String(), resp.MyThreads[0].Messages[0].ID)
	gs.Equal(f.Messages[0].Content, resp.MyThreads[0].Messages[0].Content)
	gs.Equal(f.Users[1].Nickname, resp.MyThreads[0].Messages[0].Sender.Nickname)

	thread := f.Threads[0]
	err = thread.Load("Participants")
	gs.NoError(err)

	participants, err := thread.GetParticipants()
	gs.NoError(err)
	gs.Equal(2, len(participants), "incorrect number of thread participants")

	gs.Equal(participants[0].Nickname, resp.MyThreads[0].Participants[0].Nickname)
	gs.Equal(participants[1].Nickname, resp.MyThreads[0].Participants[1].Nickname)
}
