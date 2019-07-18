package grifts

import (
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/markbates/grift/grift"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	_ = grift.Add("seed", func(c *grift.Context) error {

		organizationUuid1 := domain.GetUuidAsString()
		organizationUuid2 := domain.GetUuidAsString()
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
				err = fmt.Errorf("error loading organization fixture ... %+v\n %v", org, err.Error())
				return err
			}
		}

		userUuid1 := domain.GetUuidAsString()
		userUuid2 := domain.GetUuidAsString()
		userUuid3 := domain.GetUuidAsString()
		userUuid4 := domain.GetUuidAsString()
		userUuid5 := domain.GetUuidAsString()
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
				err = fmt.Errorf("error loading user fixture ... %+v\n %v", user, err.Error())
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

		postUuid1 := domain.GetUuid()
		postUuid2 := domain.GetUuid()
		postUuid3 := domain.GetUuid()
		postUuid4 := domain.GetUuid()
		postUuid5 := domain.GetUuid()
		fixturePosts := []*models.Post{
			{
				ID:             1,
				CreatedByID:    1,
				Type:           "request",
				OrganizationID: 1,
				Status:         "unfulfilled",
				Title:          "Maple Syrup",
				Destination:    nulls.NewString("Madrid, Spain"),
				Size:           "Medium",
				Uuid:           postUuid1,
				ReceiverID:     nulls.NewInt(1),
				NeededAfter:    time.Date(2019, time.July, 19, 0, 0, 0, 0, time.UTC),
				NeededBefore:   time.Date(2019, time.August, 3, 0, 0, 0, 0, time.UTC),
				Category:       "Unknown",
				Description:    nulls.NewString("Missing my good, old, Canadian maple syrupy goodness"),
			},
			{
				ID:             2,
				CreatedByID:    2,
				Type:           "request",
				OrganizationID: 1,
				Status:         "unfulfilled",
				Title:          "Jif Peanut Butter",
				Destination:    nulls.NewString("JAARS, NC, USA"),
				Size:           "Small",
				Uuid:           postUuid2,
				ReceiverID:     nulls.NewInt(2),
				NeededBefore:   time.Date(2019, time.August, 3, 0, 0, 0, 0, time.UTC),
				Category:       "Food",
				Description:    nulls.NewString("Jiffy Peanut Butter goes on our daily bread!"),
			},
			{
				ID:             3,
				CreatedByID:    3,
				Type:           "request",
				OrganizationID: 1,
				Status:         "unfulfilled",
				Title:          "Burt's Bee's Lip Balm",
				Destination:    nulls.NewString("Atlanta, GA, USA"),
				Size:           "Tiny",
				Uuid:           postUuid3,
				ReceiverID:     nulls.NewInt(3),
				NeededAfter:    time.Date(2019, time.July, 18, 0, 0, 0, 0, time.UTC),
				Category:       "Personal",
				Description:    nulls.NewString("Please save me from having painfully cracked lips!"),
			},
			{
				ID:             4,
				CreatedByID:    4,
				Type:           "request",
				OrganizationID: 1,
				Status:         "unfulfilled",
				Title:          "Peanut Butter",
				Destination:    nulls.NewString("Orlando, FL, USA"),
				Size:           "Small",
				Uuid:           postUuid4,
				ReceiverID:     nulls.NewInt(4),
				NeededAfter:    time.Date(2019, time.August, 3, 0, 0, 0, 0, time.UTC),
				NeededBefore:   time.Date(2019, time.September, 1, 0, 0, 0, 0, time.UTC),
				Category:       "Food",
				Description:    nulls.NewString("I already have chocolate, but I need peanut butter."),
			},
			{
				ID:             5,
				CreatedByID:    5,
				Type:           "request",
				OrganizationID: 2,
				Status:         "unfulfilled",
				Title:          "Altoids",
				Size:           "Tiny",
				Uuid:           postUuid5,
				ReceiverID:     nulls.NewInt(5),
				Category:       "Mints",
				Description:    nulls.NewString("The original celebrated curiously strong mints"),
			},
		}

		for _, post := range fixturePosts {
			err := models.DB.Create(post)
			if err != nil {
				err = fmt.Errorf("error loading post fixture ... %+v\n %v", post, err.Error())
				return err
			}
		}

		return nil
	})

})
