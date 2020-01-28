package models

type providersFixtures struct {
	Users
	Posts
	PotentialProviders
}

func createProvidersFixtures(ms *ModelSuite) providersFixtures {
	uf := createUserFixtures(ms.DB, 4)
	posts := createPostFixtures(ms.DB, 2, 0, false)
	providers := PotentialProviders{}

	for i, p := range posts {
		for _, u := range uf.Users[i+1:] {
			c := PotentialProvider{PostID: p.ID, UserID: u.ID}
			c.Create()
			providers = append(providers, c)
		}
	}

	return providersFixtures{
		Users:              uf.Users,
		Posts:              posts,
		PotentialProviders: providers,
	}
}
