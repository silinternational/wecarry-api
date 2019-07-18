package grifts

import (
	uuid2 "github.com/gofrs/uuid"
	"github.com/markbates/grift/grift"
	"github.com/silinternational/handcarry-api/models"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	_ = grift.Add("seed", func(c *grift.Context) error {

		organizationUuid1 := getUuid()
		organizationUuid2 := getUuid()
		fixtureOrgs := []*models.Organization{
			{
				ID:         1,
				Uuid:       organizationUuid1,
				Name:       "AppsDev",
				AuthType:   "SAML",
				AuthConfig: "{}",
			},
			{
				ID:         2,
				Uuid:       organizationUuid2,
				Name:       "Other",
				AuthType:   "SAML",
				AuthConfig: "{}",
			},
		}

		for _, org := range fixtureOrgs {
			err := models.DB.Create(org)
			if err != nil {
				return err
			}
		}

		userUuid1 := getUuid()
		userUuid2 := getUuid()
		userUuid3 := getUuid()
		userUuid4 := getUuid()
		userUuid5 := getUuid()
		fixtureUsers := []*models.User{
			{
				ID:         1,
				Uuid:       userUuid1,
				Email:      "clark.kent@example.org",
				FirstName:  "Clark",
				LastName:   "Kent",
				Nickname:   "Reporter38",
				AuthOrgID:  1,
				AuthOrgUid: "clark_kent",
			},
			{
				ID:         2,
				Uuid:       userUuid2,
				Email:      "jane.eyre@example.org",
				FirstName:  "Jane",
				LastName:   "Eyre",
				Nickname:   "Charlotte47",
				AuthOrgID:  1,
				AuthOrgUid: "jane_eyre",
			},
			{
				ID:         3,
				Uuid:       userUuid3,
				Email:      "jane.doe@example.org",
				FirstName:  "Jane",
				LastName:   "Doe",
				Nickname:   "Unknown42",
				AuthOrgID:  1,
				AuthOrgUid: "jane_doe",
			},
			{
				ID:         4,
				Uuid:       userUuid4,
				Email:      "denethor.ben.ecthelion@example.org",
				FirstName:  "Denethor",
				LastName:   "",
				Nickname:   "Gondor2930",
				AuthOrgID:  1,
				AuthOrgUid: "denethor",
			},
			{
				ID:         5,
				Uuid:       userUuid5,
				Email:      "john.smith@example.org",
				FirstName:  "John",
				LastName:   "Smith",
				Nickname:   "Highway1991",
				AuthOrgID:  2,
				AuthOrgUid: "john_smith",
			},
		}

		for _, user := range fixtureUsers {
			err := models.DB.Create(user)
			if err != nil {
				return err
			}
		}

		fixtureUserOrgs := []*models.UserOrganization{
			{
				ID:             1,
				OrganizationID: 1,
				UserID:         1,
				Role:           "admin",
			},
			{
				ID:             2,
				OrganizationID: 1,
				UserID:         2,
				Role:           "foo",
			},
			{
				ID:             3,
				OrganizationID: 1,
				UserID:         3,
				Role:           "bar",
			},
			{
				ID:             4,
				OrganizationID: 1,
				UserID:         4,
				Role:           "baz",
			},
			{
				ID:             5,
				OrganizationID: 2,
				UserID:         5,
				Role:           "admin",
			},
		}

		for _, userOrgs := range fixtureUserOrgs {
			err := models.DB.Create(userOrgs)
			if err != nil {
				return err
			}
		}

		return nil
	})

})

func getUuid() string {
	uuid, _ := uuid2.NewV4()
	return uuid.String()
}
