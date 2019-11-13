package models

import (
	"github.com/silinternational/wecarry-api/domain"
)

type OrganizationFixtures struct {
	Organizations
	OrganizationDomains
	Users
}

func CreateFixturesForOrganizationGetDomains(ms *ModelSuite) OrganizationFixtures {
	org := Organization{
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	createFixture(ms, &org)

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

	user := User{
		Email:     "user1@example.com",
		FirstName: "Existing",
		LastName:  "User",
		Nickname:  "Existing User ",
		Uuid:      domain.GetUuid(),
	}
	createFixture(ms, &user)

	return OrganizationFixtures{
		Organizations:       Organizations{org},
		OrganizationDomains: orgDomains,
		Users:               Users{user},
	}
}
