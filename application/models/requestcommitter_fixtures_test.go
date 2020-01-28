package models

type committersFixtures struct {
	Users
	Posts
	RequestCommitters
}

func createCommittersFixtures(ms *ModelSuite) committersFixtures {
	uf := createUserFixtures(ms.DB, 4)
	posts := createPostFixtures(ms.DB, 2, 0, false)
	committers := RequestCommitters{}

	for i, p := range posts {
		for _, u := range uf.Users[i+1:] {
			c := RequestCommitter{PostID: p.ID, UserID: u.ID}
			c.Create()
			committers = append(committers, c)
		}
	}

	return committersFixtures{
		Users:             uf.Users,
		Posts:             posts,
		RequestCommitters: committers,
	}
}
