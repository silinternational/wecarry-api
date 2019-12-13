package models

type OrganizationFixtures struct {
	Organizations
	OrganizationDomains
	Users
}

func createFixturesForOrganizationGetUsers(ms *ModelSuite) OrganizationFixtures {
	uf := CreateUserFixtures(ms.DB, 3)
	org := uf.Organization
	users := uf.Users

	// nicknames in unsorted order
	nicknames := []string{"alice", "john", "bob"}
	for i := range users {
		users[i].Nickname = nicknames[i]
		ms.NoError(ms.DB.Save(&users[i]))
	}

	return OrganizationFixtures{
		Organizations: Organizations{org},
		Users:         users,
	}
}

func CreateFixturesForOrganizationGetDomains(ms *ModelSuite) OrganizationFixtures {
	uf := CreateUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	orgDomains := OrganizationDomains{
		{
			OrganizationID: org.ID,
			Domain:         "example.org",
		},
		{
			OrganizationID: org.ID,
			Domain:         "1.example.org",
		},
		{
			OrganizationID: org.ID,
			Domain:         "example.com",
		},
	}
	for i := range orgDomains {
		createFixture(ms, &orgDomains[i])
	}

	return OrganizationFixtures{
		Organizations:       Organizations{org},
		OrganizationDomains: orgDomains,
		Users:               Users{user},
	}
}
