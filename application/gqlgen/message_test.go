package gqlgen

type messageResponse struct {
	Message struct {
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
		Thread struct {
			ID           string `json:"id"`
			Participants []struct {
				ID       string `json:"id"`
				Nickname string `json:"nickname"`
			} `json:"participants"`
		} `json:"thread"`
	} `json:"message"`
}

func (gs *GqlgenSuite) TestMessageQuery() {
	f := createFixtures_MessageQuery(gs)
	c := getGqlClient()

	query := `{ message(id: "` + f.Messages[0].Uuid.String() + `")
		{ id content sender { location { description }} thread {id participants {nickname}}}}`

	var resp messageResponse

	TestUser = f.Users[0]
	err := c.Post(query, &resp)
	gs.NoError(err)

	thread, err := f.Messages[0].GetThread()
	gs.NoError(err)

	participants, err := thread.GetParticipants()
	gs.NoError(err)
	gs.Equal(2, len(participants), "incorrect number of thread participants")

	gs.Equal(f.Messages[0].Uuid.String(), resp.Message.ID)
	gs.Equal(f.Messages[0].Content, resp.Message.Content)
	gs.Equal(f.Messages[0].SentBy.Nickname, resp.Message.Sender.Nickname)
	gs.Equal(f.Messages[0].SentBy.Location.Description, resp.Message.Sender.Location.Description)
	gs.Equal(thread.Uuid.String(), resp.Message.Thread.ID)
	gs.Equal(participants[0].Nickname, resp.Message.Thread.Participants[0].Nickname)
	gs.Equal(participants[1].Nickname, resp.Message.Thread.Participants[1].Nickname)
}

func (gs *GqlgenSuite) TestCreateMessage() {
	f := createFixtures_MessageQuery(gs)
	c := getGqlClient()

	newContent := "New Message Created"

	input := `postID: "` + f.Posts[0].Uuid.String() + `" ` +
		`threadID: "` + f.Threads[0].Uuid.String() + `" content: "` + newContent + `" `
	query := `mutation { message: createMessage (input: {` + input +
		`}) { id content thread {id}}}`

	var resp messageResponse

	TestUser = f.Users[0]
	err := c.Post(query, &resp)
	gs.NoError(err)

	thread, err := f.Messages[0].GetThread()
	gs.NoError(err)

	messages, err := thread.GetMessages()
	gs.NoError(err)
	gs.Equal(3, len(messages), "incorrect number of thread messages")

	gs.Equal(newContent, resp.Message.Content)
	gs.Equal(thread.Uuid.String(), resp.Message.Thread.ID)
}
