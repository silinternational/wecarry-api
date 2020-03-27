package actions

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

	thread, err := f.Messages[0].GetThread()
	as.NoError(err)

	participants, err := thread.GetParticipants()
	as.NoError(err)
	as.Equal(2, len(participants), "incorrect number of thread participants")

	as.Equal(f.Messages[0].UUID.String(), resp.Message.ID)
	as.Equal(f.Messages[0].Content, resp.Message.Content)
	as.Equal(f.Users[1].Nickname, resp.Message.Sender.Nickname)
	as.Equal(thread.UUID.String(), resp.Message.Thread.ID)
	as.Equal(participants[0].Nickname, resp.Message.Thread.Participants[0].Nickname)
	as.Equal(participants[1].Nickname, resp.Message.Thread.Participants[1].Nickname)
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

	thread, err := f.Messages[0].GetThread()
	as.NoError(err)

	messages, err := thread.Messages()
	as.NoError(err)
	as.Equal(3, len(messages), "incorrect number of thread messages")

	as.Equal(newContent, resp.Message.Content)
	as.Equal(thread.UUID.String(), resp.Message.Thread.ID)
}
