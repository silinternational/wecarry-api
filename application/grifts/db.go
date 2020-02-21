package grifts

import (
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/markbates/grift/grift"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	_ = grift.Add("seed", func(c *grift.Context) error {

		var existingOrgs models.Organizations
		_ = models.DB.All(&existingOrgs)
		if len(existingOrgs) > 1 {
			return nil
		}

		// ORGANIZATIONS Table
		organizationUUID2, _ := uuid.FromString("d2e95724-9270-4050-82d9-6a9f9c35c766")
		fixtureOrgs := []*models.Organization{
			{
				UUID:       organizationUUID2,
				Name:       "Other",
				AuthType:   models.AuthTypeSaml,
				AuthConfig: "{}",
			},
		}

		for i, org := range fixtureOrgs {
			err := models.DB.Create(fixtureOrgs[i])
			if err != nil {
				err = fmt.Errorf("error loading organization fixture ... %+v\n %v", org, err.Error())
				return err
			}
		}

		var primaryOrgID int
		if len(existingOrgs) > 0 {
			primaryOrgID = existingOrgs[0].ID
		} else {
			primaryOrgID = fixtureOrgs[0].ID
		}

		// USERS Table
		userUUID1, _ := uuid.FromString("e5447366-26b2-4256-b2ab-58c92c3d54cc")
		userUUID2, _ := uuid.FromString("3d79902f-c204-4922-b479-57f0ec41eabe")
		userUUID3, _ := uuid.FromString("babcf980-e1f0-42d3-b2b0-2e4704159f4f")
		userUUID4, _ := uuid.FromString("44dc63fa-1227-4bea-b34a-416a26c3e077")
		userUUID5, _ := uuid.FromString("2a96a5a6-971a-403d-8276-c41657bc57ce")
		fixtureUsers := []*models.User{
			{
				UUID:      userUUID1,
				Email:     "clark.kent@example.org",
				FirstName: "Clark",
				LastName:  "Kent",
				Nickname:  "Reporter38",
			},
			{
				UUID:      userUUID2,
				Email:     "jane.eyre@example.org",
				FirstName: "Jane",
				LastName:  "Eyre",
				Nickname:  "Charlotte47",
			},
			{
				UUID:      userUUID3,
				Email:     "jane.doe@example.org",
				FirstName: "Jane",
				LastName:  "Doe",
				Nickname:  "Unknown42",
			},
			{
				UUID:      userUUID4,
				Email:     "denethor.ben.ecthelion@example.org",
				FirstName: "Denethor",
				LastName:  "",
				Nickname:  "Gondor2930",
			},
			{
				UUID:      userUUID5,
				Email:     "john.smith@example.org",
				FirstName: "John",
				LastName:  "Smith",
				Nickname:  "Highway1991",
			},
		}

		for i, user := range fixtureUsers {
			err := models.DB.Create(fixtureUsers[i])
			if err != nil {
				err = fmt.Errorf("error loading user fixture ... %+v\n %v", user, err.Error())
				return err
			}
		}

		// USER_ORGANIZATIONS Table
		fixtureUserOrgs := []*models.UserOrganization{
			{
				OrganizationID: primaryOrgID,
				UserID:         fixtureUsers[0].ID,
				Role:           models.UserOrganizationRoleAdmin,
				AuthEmail:      "clark.kent@example.org",
				AuthID:         "clark_kent",
			},
			{
				OrganizationID: primaryOrgID,
				UserID:         fixtureUsers[0].ID,
				Role:           models.UserOrganizationRoleUser,
				AuthEmail:      "jane.eyre@example.org",
				AuthID:         "jane_eyre",
			},
			{
				OrganizationID: primaryOrgID,
				UserID:         fixtureUsers[0].ID,
				Role:           models.UserOrganizationRoleUser,
				AuthEmail:      "jane.doe@example.org",
				AuthID:         "jane_doe",
			},
			{
				OrganizationID: primaryOrgID,
				UserID:         fixtureUsers[0].ID,
				Role:           models.UserOrganizationRoleUser,
				AuthEmail:      "denethor.ben.ecthelion@example.org",
				AuthID:         "denethor_ecthelion",
			},
			{
				OrganizationID: fixtureOrgs[0].ID,
				UserID:         fixtureUsers[0].ID,
				Role:           models.UserOrganizationRoleAdmin,
				AuthEmail:      "john.smith@example.org",
				AuthID:         "john_smith",
			},
		}

		for i, userOrgs := range fixtureUserOrgs {
			fixtureUserOrgs[i].UserID = fixtureUsers[i].ID
			err := models.DB.Create(fixtureUserOrgs[i])
			if err != nil {
				err = fmt.Errorf("error loading user organizations fixture ... %+v\n %v", userOrgs, err.Error())
				return err
			}
		}

		// LOCATIONS Table
		fixtureLocations := []*models.Location{
			{
				Description: "Madrid, Spain",
				Country:     "ES",
				Latitude:    nulls.NewFloat64(40.4168),
				Longitude:   nulls.NewFloat64(-3.7038),
			},
			{
				Description: "JAARS, NC, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(34.8638),
				Longitude:   nulls.NewFloat64(-80.7459),
			},
			{
				Description: "Atlanta, GA, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(33.7490),
				Longitude:   nulls.NewFloat64(-84.3880),
			},
			{
				Description: "Orlando, FL, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(28.5383),
				Longitude:   nulls.NewFloat64(-81.3792),
			},
			{
				Description: "Toronto, Canada",
				Country:     "CA",
				Latitude:    nulls.NewFloat64(43.6532),
				Longitude:   nulls.NewFloat64(-79.3832),
			},
			{
				Description: "Nairobi, Kenya",
				Country:     "KE",
				Latitude:    nulls.NewFloat64(-1.2921),
				Longitude:   nulls.NewFloat64(36.8219),
			},
			{
				Description: "Brasília, Brazil",
				Country:     "BR",
				Latitude:    nulls.NewFloat64(-15.8267),
				Longitude:   nulls.NewFloat64(-47.9218),
			},
			{
				Description: "Chiang Mai, Thailand",
				Country:     "TH",
				Latitude:    nulls.NewFloat64(18.7953),
				Longitude:   nulls.NewFloat64(98.9620),
			},
			{
				Description: "Milwaukee, WI, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(43.0389),
				Longitude:   nulls.NewFloat64(-87.9065),
			},
		}

		for i, loc := range fixtureLocations {
			err := models.DB.Create(fixtureLocations[i])
			if err != nil {
				err = fmt.Errorf("error loading locations fixture ... %+v\n %v", loc, err.Error())
				return err
			}
		}

		// POSTS Table
		futureDate := time.Now().Add(8 * domain.DurationWeek)
		postUUID1, _ := uuid.FromString("270fa549-65f2-43c0-ac27-78a054cf49a1")
		postUUID2, _ := uuid.FromString("028164cd-a8f5-43b9-98d0-f8a7778ea2f1")
		postUUID3, _ := uuid.FromString("e625a482-c8ff-4f52-b8ed-73e6b3eac4d7")
		postUUID4, _ := uuid.FromString("8e08011d-bd5f-4c1a-a4f4-0c019beb939b")
		postUUID5, _ := uuid.FromString("35e2b332-a968-4932-b205-ca0d1eabdf0e")
		fixturePosts := []*models.Post{
			{
				Type:           models.PostTypeRequest,
				OrganizationID: primaryOrgID,
				Title:          "Maple Syrup",
				Size:           models.PostSizeMedium,
				UUID:           postUUID1,
				Description:    nulls.NewString("Missing my good, old, Canadian maple syrupy goodness"),
			},
			{
				Type:           models.PostTypeRequest,
				OrganizationID: primaryOrgID,
				Title:          "Jif Peanut Butter",
				Size:           models.PostSizeSmall,
				UUID:           postUUID2,
				Description:    nulls.NewString("Jiffy Peanut Butter goes on our daily bread!"),
			},
			{
				Type:           models.PostTypeRequest,
				OrganizationID: primaryOrgID,
				Title:          "Burt's Bee's Lip Balm",
				Size:           models.PostSizeTiny,
				UUID:           postUUID3,
				Description:    nulls.NewString("Please save me from having painfully cracked lips!"),
			},
			{
				Type:           models.PostTypeRequest,
				OrganizationID: primaryOrgID,
				Title:          "Peanut Butter",
				Size:           models.PostSizeSmall,
				UUID:           postUUID4,
				Description:    nulls.NewString("I already have chocolate, but I need peanut butter."),
			},
			{
				Type:           models.PostTypeRequest,
				OrganizationID: fixtureOrgs[0].ID,
				Title:          "Altoids",
				Size:           models.PostSizeTiny,
				UUID:           postUUID5,
				Description:    nulls.NewString("The original celebrated curiously strong mints"),
			},
		}

		for i, post := range fixturePosts {
			fixturePosts[i].DestinationID = fixtureLocations[i].ID
			fixturePosts[i].Status = models.PostStatusOpen
			fixturePosts[i].CreatedByID = fixtureUsers[i].ID
			fixturePosts[i].ReceiverID = nulls.NewInt(fixtureUsers[i].ID)
			fixturePosts[i].NeededBefore = nulls.NewTime(futureDate)
			err := models.DB.Create(fixturePosts[i])
			if err != nil {
				err = fmt.Errorf("error loading post fixture ... %+v\n %v", post, err.Error())
				return err
			}
		}

		// THREADS Table
		threadUUID1, _ := uuid.FromString("bdb7515d-06a9-4896-97a4-aeae962b85e2")
		threadUUID2, _ := uuid.FromString("216c4b08-a4b4-4b7f-b62c-543be07e07c0")
		threadUUID3, _ := uuid.FromString("79adc9bf-69b6-4b8a-ae23-dc26fb9de661")
		threadUUID4, _ := uuid.FromString("7781642d-50d0-43da-9af2-e21133b4af91")
		fixtureThreads := []*models.Thread{
			{
				UUID: threadUUID1,
			},
			{
				UUID: threadUUID2,
			},
			{
				UUID: threadUUID3,
			},
			{
				UUID: threadUUID4,
			},
		}

		for i, thread := range fixtureThreads {
			fixtureThreads[i].PostID = fixturePosts[i].ID
			err := models.DB.Create(fixtureThreads[i])
			if err != nil {
				err = fmt.Errorf("error loading thread fixture ... %+v\n %v", thread, err.Error())
				return err
			}
		}

		// THREAD_PARTICIPANTS Table
		fixtureParticipants := []*models.ThreadParticipant{
			{
				ThreadID: fixtureThreads[0].ID,
				UserID:   fixtureUsers[0].ID,
			},
			{
				ThreadID: fixtureThreads[0].ID,
				UserID:   fixtureUsers[4].ID,
			},
			{
				ThreadID: fixtureThreads[1].ID,
				UserID:   fixtureUsers[1].ID,
			},
			{
				ThreadID: fixtureThreads[1].ID,
				UserID:   fixtureUsers[4].ID,
			},
			{
				ThreadID: fixtureThreads[2].ID,
				UserID:   fixtureUsers[2].ID,
			},
			{
				ThreadID: fixtureThreads[2].ID,
				UserID:   fixtureUsers[4].ID,
			},
			{
				ThreadID: fixtureThreads[3].ID,
				UserID:   fixtureUsers[3].ID,
			},
		}

		for i, participant := range fixtureParticipants {
			err := models.DB.Create(fixtureParticipants[i])
			if err != nil {
				err = fmt.Errorf("error loading thread participant fixture ... %+v\n %v", participant, err.Error())
				return err
			}
		}

		// MESSAGES Table
		messageUUID1, _ := uuid.FromString("b0d7c515-e74c-4af7-a937-f1deb9369831")
		messageUUID2, _ := uuid.FromString("ac52793a-e683-4684-bc10-213f49a3e302")
		messageUUID3, _ := uuid.FromString("b90703f6-a5d7-4534-aacd-6b3212288454")
		messageUUID4, _ := uuid.FromString("a74c0cb6-66e6-43d4-9c71-0ce96bdda99b")
		messageUUID5, _ := uuid.FromString("e3932ab7-ae53-493f-a676-50512c4ca952")
		messageUUID6, _ := uuid.FromString("0aea9161-b374-45ae-8abd-faf04b8da9e1")
		messageUUID7, _ := uuid.FromString("d9e54392-1a5f-4e6e-b74a-10756b8a9812")
		fixtureMessages := []*models.Message{
			{
				ThreadID: fixtureThreads[0].ID,
				UUID:     messageUUID1,
				SentByID: fixtureUsers[4].ID,
				Content:  "Any chance you can bring some PB?",
			},
			{
				ThreadID: fixtureThreads[0].ID,
				UUID:     messageUUID2,
				SentByID: fixtureUsers[0].ID,
				Content:  "Absolutely!",
			},
			{
				ThreadID: fixtureThreads[0].ID,
				UUID:     messageUUID3,
				SentByID: fixtureUsers[4].ID,
				Content:  "Thanks 😁",
			},
			{
				ThreadID: fixtureThreads[1].ID,
				UUID:     messageUUID4,
				SentByID: fixtureUsers[4].ID,
				Content:  "red plum jam, if possible",
			},
			{
				ThreadID: fixtureThreads[2].ID,
				UUID:     messageUUID5,
				SentByID: fixtureUsers[2].ID,
				Content:  "Did you find any Wintergreen Altoids?",
			},
			{
				ThreadID: fixtureThreads[2].ID,
				UUID:     messageUUID6,
				SentByID: fixtureUsers[4].ID,
				Content:  "No luck, sorry",
			},
			{
				ThreadID: fixtureThreads[3].ID,
				UUID:     messageUUID7,
				SentByID: fixtureUsers[3].ID,
				Content:  "I haven't heard from my son, either. Have you seen him recently?",
			},
		}

		for i, message := range fixtureMessages {
			err := models.DB.Create(fixtureMessages[i])
			if err != nil {
				err = fmt.Errorf("error loading message fixture ... %+v\n %v", message, err.Error())
				return err
			}
		}

		// FILES Table
		fileUUID1, _ := uuid.FromString("a7103b02-9b50-49a1-9776-b63f1cb7e84b")
		fileUUID2, _ := uuid.FromString("a74569e6-fd54-4945-a9de-c05c711938ee")
		fileUUID3, _ := uuid.FromString("c1eed7f0-2c8f-4d17-911c-99a54d29b0a1")
		fixtureFiles := []*models.File{
			{
				UUID:          fileUUID1,
				Name:          "iccmlogo.png",
				URL:           "https://iccm.africa/img/iccmlogo.png",
				URLExpiration: time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC),
				Size:          5279,
				ContentType:   "image/png",
			},
			{
				UUID:          fileUUID2,
				Name:          "thesend.png",
				URL:           "http://thesend.org.br/wp-content/uploads/2019/06/logo.png",
				URLExpiration: time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC),
				Size:          15,
				ContentType:   "image/png",
			},
			{
				UUID:          fileUUID3,
				Name:          "logo.png",
				URL:           "https://static.wixstatic.com/media/f85009_a5b1c807a4a34e3284576f8e0cf334ca~mv2.jpg/v1/fill/w_755,h_1008,al_c,q_85/f85009_a5b1c807a4a34e3284576f8e0cf334ca~mv2.jpg",
				URLExpiration: time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC),
				Size:          15000,
				ContentType:   "image/jpg",
			},
		}

		for i, file := range fixtureFiles {
			err := models.DB.Create(fixtureFiles[i])
			if err != nil {
				err = fmt.Errorf("error loading file fixture ... %+v\n %v", file, err.Error())
				return err
			}
		}

		// MEETINGS Table
		meetingUUID1, _ := uuid.FromString("4a747184-7c7e-426f-a1fc-c310428f3d8d")
		meetingUUID2, _ := uuid.FromString("48b9279f-d31f-41f5-8926-c5231c278aaa")
		meetingUUID3, _ := uuid.FromString("7beac887-a03f-436f-addb-16a2e2880d67")
		meetingUUID4, _ := uuid.FromString("d5c8185e-a084-494f-a549-adb73f3686ee")
		fixtureMeetings := []*models.Meeting{
			{
				UUID:        meetingUUID1,
				CreatedByID: fixtureUsers[0].ID,
				Name:        "IT Connect / ICCM",
				MoreInfoURL: nulls.NewString("https://iccm.africa"),
				LocationID:  fixtureLocations[5].ID,
				ImageFileID: nulls.NewInt(fixtureFiles[0].ID),
				StartDate:   time.Date(2020, 3, 15, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2020, 3, 21, 0, 0, 0, 0, time.UTC),
				InviteCode:  nulls.NewUUID(meetingUUID1),
			},
			{
				UUID:        meetingUUID2,
				CreatedByID: fixtureUsers[0].ID,
				Name:        "The Send Brazil",
				MoreInfoURL: nulls.NewString("http://thesend.org.br/en-2/"),
				LocationID:  fixtureLocations[6].ID,
				ImageFileID: nulls.NewInt(fixtureFiles[1].ID),
				StartDate:   time.Date(2021, 2, 8, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2021, 2, 8, 0, 0, 0, 0, time.UTC),
				InviteCode:  nulls.NewUUID(meetingUUID2),
			},
			{
				UUID:        meetingUUID3,
				CreatedByID: fixtureUsers[4].ID,
				Name:        "ICON20",
				LocationID:  fixtureLocations[7].ID,
				StartDate:   time.Date(2020, 4, 4, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2020, 4, 9, 0, 0, 0, 0, time.UTC),
			},
			{
				UUID:        meetingUUID4,
				CreatedByID: fixtureUsers[2].ID,
				Name:        "Fresh Fish Suppliers of America",
				LocationID:  fixtureLocations[8].ID,
				ImageFileID: nulls.NewInt(fixtureFiles[2].ID),
				StartDate:   time.Date(2020, 4, 4, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2020, 4, 9, 0, 0, 0, 0, time.UTC),
			}}

		for i, meeting := range fixtureMeetings {
			err := models.DB.Create(fixtureMeetings[i])
			if err != nil {
				err = fmt.Errorf("error loading meeting fixture ... %+v\n %v", meeting, err.Error())
				return err
			}
		}

		// meeting_invites table
		inviteSecret1, _ := uuid.FromString("ad08446a-65dc-4a31-9c67-497dace2d519")
		inviteSecret2, _ := uuid.FromString("7351594a-cf3a-4b5c-b133-1f5e029e8e18")
		inviteSecret3, _ := uuid.FromString("5ef7e8e4-33da-4fa8-a053-1570114018d8")
		inviteSecret4, _ := uuid.FromString("5521a5cf-83a4-45b9-a579-83f668cee97e")
		fixtureInvites := []*models.MeetingInvite{
			{
				MeetingID: fixtureMeetings[0].ID,
				InviterID: fixtureUsers[0].ID,
				Secret:    inviteSecret1,
				Email:     "clark.kent@example.org",
			},
			{
				MeetingID: fixtureMeetings[0].ID,
				InviterID: fixtureUsers[0].ID,
				Secret:    inviteSecret2,
				Email:     "elmer_fudd@example.org",
			},
			{
				MeetingID: fixtureMeetings[0].ID,
				InviterID: fixtureUsers[0].ID,
				Secret:    inviteSecret3,
				Email:     "another.yahoo@example.com",
			},
			{
				MeetingID: fixtureMeetings[0].ID,
				InviterID: fixtureUsers[0].ID,
				Secret:    inviteSecret4,
				Email:     "jimmy-crack-corn@example.net",
			},
		}

		for i, meeting := range fixtureInvites {
			err := models.DB.Create(fixtureInvites[i])
			if err != nil {
				err = fmt.Errorf("error loading invite fixture ... %+v\n %v", meeting, err.Error())
				return err
			}
		}

		// meeting_participants table
		fixtureMeetingParticipants := []*models.MeetingParticipant{
			{
				MeetingID:   fixtureMeetings[0].ID,
				UserID:      fixtureUsers[0].ID,
				InviteID:    nulls.NewInt(fixtureInvites[0].ID),
				IsOrganizer: true,
			},
			{
				MeetingID:   fixtureMeetings[0].ID,
				UserID:      fixtureUsers[1].ID,
				IsOrganizer: false,
			},
		}

		for i, meeting := range fixtureMeetingParticipants {
			err := models.DB.Create(fixtureMeetingParticipants[i])
			if err != nil {
				err = fmt.Errorf("error loading participant fixture ... %+v\n %v", meeting, err.Error())
				return err
			}
		}

		return nil
	})

})
