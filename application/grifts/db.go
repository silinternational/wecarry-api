package grifts

import (
	"fmt"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/markbates/grift/grift"
	"github.com/silinternational/handcarry-api/models"
	"time"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	_ = grift.Add("seed", func(c *grift.Context) error {

		// ORGANIZATIONS Table
		organizationUuid1 := "f3a79b30-f00e-48a0-a64d-e27748dea22d"
		organizationUuid2 := "d2e95724-9270-4050-82d9-6a9f9c35c766"
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

		// USERS Table
		userUuid1, _ := uuid.FromString("e5447366-26b2-4256-b2ab-58c92c3d54cc")
		userUuid2, _ := uuid.FromString("3d79902f-c204-4922-b479-57f0ec41eabe")
		userUuid3, _ := uuid.FromString("babcf980-e1f0-42d3-b2b0-2e4704159f4f")
		userUuid4, _ := uuid.FromString("44dc63fa-1227-4bea-b34a-416a26c3e077")
		userUuid5, _ := uuid.FromString("2a96a5a6-971a-403d-8276-c41657bc57ce")
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

		// USER_ORGANIZATIONS Table
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
				err = fmt.Errorf("error loading user organizations fixture ... %+v\n %v", userOrgs, err.Error())
				return err
			}
		}

		// POSTS Table
		postUuid1, _ := uuid.FromString("270fa549-65f2-43c0-ac27-78a054cf49a1")
		postUuid2, _ := uuid.FromString("028164cd-a8f5-43b9-98d0-f8a7778ea2f1")
		postUuid3, _ := uuid.FromString("e625a482-c8ff-4f52-b8ed-73e6b3eac4d7")
		postUuid4, _ := uuid.FromString("8e08011d-bd5f-4c1a-a4f4-0c019beb939b")
		postUuid5, _ := uuid.FromString("35e2b332-a968-4932-b205-ca0d1eabdf0e")
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

		// THREADS Table
		threadUuid1, _ := uuid.FromString("bdb7515d-06a9-4896-97a4-aeae962b85e2")
		threadUuid2, _ := uuid.FromString("216c4b08-a4b4-4b7f-b62c-543be07e07c0")
		threadUuid3, _ := uuid.FromString("79adc9bf-69b6-4b8a-ae23-dc26fb9de661")
		threadUuid4, _ := uuid.FromString("7781642d-50d0-43da-9af2-e21133b4af91")
		fixtureThreads := []*models.Thread{
			{
				ID:     1,
				Uuid:   threadUuid1,
				PostID: 1,
			},
			{
				ID:     2,
				Uuid:   threadUuid2,
				PostID: 2,
			},
			{
				ID:     3,
				Uuid:   threadUuid3,
				PostID: 3,
			},
			{
				ID:     4,
				Uuid:   threadUuid4,
				PostID: 4,
			},
		}

		for _, thread := range fixtureThreads {
			err := models.DB.Create(thread)
			if err != nil {
				err = fmt.Errorf("error loading thread fixture ... %+v\n %v", thread, err.Error())
				return err
			}
		}

		// THREAD_PARTICIPANTS Table
		fixtureParticipants := []*models.ThreadParticipant{
			{
				ID:       1,
				ThreadID: 1,
				UserID:   1,
			},
			{
				ID:       2,
				ThreadID: 1,
				UserID:   5,
			},
			{
				ID:       3,
				ThreadID: 2,
				UserID:   2,
			},
			{
				ID:       4,
				ThreadID: 2,
				UserID:   5,
			},
			{
				ID:       5,
				ThreadID: 3,
				UserID:   3,
			},
			{
				ID:       6,
				ThreadID: 3,
				UserID:   5,
			},
			{
				ID:       7,
				ThreadID: 4,
				UserID:   4,
			},
		}

		for _, participant := range fixtureParticipants {
			err := models.DB.Create(participant)
			if err != nil {
				err = fmt.Errorf("error loading thread participant fixture ... %+v\n %v", participant, err.Error())
				return err
			}
		}

		// MESSAGES Table
		messageUuid1, _ := uuid.FromString("b0d7c515-e74c-4af7-a937-f1deb9369831")
		messageUuid2, _ := uuid.FromString("ac52793a-e683-4684-bc10-213f49a3e302")
		messageUuid3, _ := uuid.FromString("b90703f6-a5d7-4534-aacd-6b3212288454")
		messageUuid4, _ := uuid.FromString("a74c0cb6-66e6-43d4-9c71-0ce96bdda99b")
		messageUuid5, _ := uuid.FromString("e3932ab7-ae53-493f-a676-50512c4ca952")
		messageUuid6, _ := uuid.FromString("0aea9161-b374-45ae-8abd-faf04b8da9e1")
		messageUuid7, _ := uuid.FromString("d9e54392-1a5f-4e6e-b74a-10756b8a9812")
		fixtureMessages := []*models.Message{
			{
				ThreadID: 1,
				ID:       1,
				Uuid:     messageUuid1,
				SentByID: 5,
				Content:  "Any chance you can bring some PB?",
			},
			{
				ThreadID: 1,
				ID:       2,
				Uuid:     messageUuid2,
				SentByID: 1,
				Content:  "Absolutely!",
			},
			{
				ThreadID: 1,
				ID:       3,
				Uuid:     messageUuid3,
				SentByID: 5,
				Content:  "Thanks 😁",
			},
			{
				ThreadID: 2,
				ID:       4,
				Uuid:     messageUuid4,
				SentByID: 5,
				Content:  "red plum jam, if possible",
			},
			{
				ThreadID: 3,
				ID:       5,
				Uuid:     messageUuid5,
				SentByID: 3,
				Content:  "Did you find any Wintergreen Altoids?",
			},
			{
				ThreadID: 3,
				ID:       6,
				Uuid:     messageUuid6,
				SentByID: 5,
				Content:  "No luck, sorry",
			},
			{
				ThreadID: 4,
				ID:       7,
				Uuid:     messageUuid7,
				SentByID: 4,
				Content:  "I haven't heard from my son, either. Have you seen him recently?",
			},
		}

		for _, message := range fixtureMessages {
			err := models.DB.Create(message)
			if err != nil {
				err = fmt.Errorf("error loading message fixture ... %+v\n %v", message, err.Error())
				return err
			}
		}

		return nil
	})

})
