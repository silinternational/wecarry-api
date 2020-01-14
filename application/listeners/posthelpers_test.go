package listeners

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

func (ms *ModelSuite) TestGetPostUsers() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_GetPostUsers(ms, t)
	users := orgUserPostFixtures.users
	posts := orgUserPostFixtures.posts

	tests := []struct {
		name          string
		id            int
		wantRequester postUser
		wantProvider  postUser
		wantErr       bool
	}{
		{name: "Request by User0 with User1 as Provider",
			id: posts[0].ID,
			wantRequester: postUser{
				Language: domain.UserPreferenceLanguageEnglish,
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
			wantProvider: postUser{
				Language: domain.UserPreferenceLanguageFrench,
				Nickname: users[1].Nickname,
				Email:    users[1].Email,
			},
		},
		{name: "Request by User0 with no Provider",
			id: posts[1].ID,
			wantRequester: postUser{
				Language: domain.UserPreferenceLanguageEnglish,
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var post models.Post
			err := post.FindByID(test.id)
			ms.NoError(err, "error finding post for test")

			postUsers := getPostUsers(post)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantRequester, postUsers.Receiver)
				ms.Equal(test.wantProvider, postUsers.Provider)
			}
		})
	}
}

func (ms *ModelSuite) TestRequestStatusUpdatedNotifications() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_RequestStatusUpdatedNotifications(ms, t)
	posts := orgUserPostFixtures.posts

	postStatusEData := models.PostStatusEventData{
		OldStatus: models.PostStatusOpen,
		NewStatus: models.PostStatusAccepted,
		PostID:    posts[0].ID,
	}

	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	// No logging message expected
	requestStatusUpdatedNotifications(posts[0], postStatusEData)

	got := buf.String()
	ms.Equal("", got, "Got an unexpected error log entry")

	buf.Reset()

	// Logging message expected about bad transition
	postStatusEData.NewStatus = models.PostStatusAccepted
	requestStatusUpdatedNotifications(posts[0], postStatusEData)

	got = buf.String()
	want := "unexpected status transition 'OPEN-ACCEPTED'"
	test.AssertStringContains(t, got, want, 45)

}

func (ms *ModelSuite) TestSendNotificationRequestFromStatus() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_sendNotificationRequestFromStatus(ms, t)
	posts := orgUserPostFixtures.posts

	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	var getT = notifications.GetEmailTemplate

	tests := []struct {
		name             string
		template         string
		post             models.Post
		eventData        models.PostStatusEventData
		sendFunction     func(senderParams)
		wantEmailsSent   int
		wantToEmail      string
		wantBodyContains string
		wantEmailNumber  int
		wantErrLog       string
	}{
		{name: "Good - Accepted to Open",
			post: posts[0],
			eventData: models.PostStatusEventData{
				OldStatus:     models.PostStatusAccepted,
				NewStatus:     models.PostStatusOpen,
				PostID:        posts[0].ID,
				OldProviderID: *models.GetIntFromNullsInt(posts[0].ProviderID),
			},
			template:         domain.MessageTemplateRequestFromAcceptedToOpen,
			sendFunction:     sendNotificationRequestFromAcceptedToOpen,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "isn't ready after all to have you fulfill",
		},
		{name: "Good - Accepted to Completed",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromAcceptedToCompleted,
			sendFunction:     sendNotificationRequestFromAcceptedOrDeliveredToCompleted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "reported that they have received",
		},
		{name: "Bad - Accepted to Completed", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromAcceptedToCompleted,
			sendFunction: sendNotificationRequestFromAcceptedOrDeliveredToCompleted,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestReceived),
		},
		{name: "Good - Accepted to Delivered",
			post:             posts[0],
			template:         domain.MessageTemplateRequestDelivered,
			sendFunction:     sendNotificationRequestFromAcceptedOrCommittedToDelivered,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "reported that they have delivered your request",
		},
		{name: "Bad - Accepted to Delivered", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromAcceptedToDelivered,
			sendFunction: sendNotificationRequestFromAcceptedOrCommittedToDelivered,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				getT(domain.MessageTemplateRequestFromAcceptedToDelivered)),
		},
		{name: "Good - Accepted to Open",
			post: posts[0],
			eventData: models.PostStatusEventData{
				OldStatus:     models.PostStatusAccepted,
				NewStatus:     models.PostStatusOpen,
				PostID:        posts[0].ID,
				OldProviderID: *models.GetIntFromNullsInt(posts[0].ProviderID),
			},
			template:         domain.MessageTemplateRequestFromAcceptedToOpen,
			sendFunction:     sendNotificationRequestFromAcceptedToOpen,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "isn't ready after all to have you fulfill",
		},
		{name: "Good - Accepted to Removed",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromAcceptedToRemoved,
			sendFunction:     sendNotificationRequestFromAcceptedToRemoved,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "has removed this request",
		},
		{name: "Bad - Accepted to Removed", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromAcceptedToRemoved,
			sendFunction: sendNotificationRequestFromAcceptedToRemoved,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestFromAcceptedToRemoved),
		},
		{name: "Good - Completed to Accepted",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromCompletedToAccepted,
			sendFunction:     sendNotificationRequestFromCompletedToAcceptedOrDelivered,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "but now have corrected that",
		},
		{name: "Bad - Completed to Accepted", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromCompletedToAccepted,
			sendFunction: sendNotificationRequestFromCompletedToAcceptedOrDelivered,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestNotReceivedAfterAll),
		},
		{name: "Good - Delivered to Accepted",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromDeliveredToAccepted,
			sendFunction:     sendNotificationRequestFromDeliveredToAccepted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "corrected themselves to say they haven't",
		},
		{name: "Good - Delivered to Completed",
			post:             posts[0],
			template:         domain.MessageTemplateRequestReceived,
			sendFunction:     sendNotificationRequestFromAcceptedOrDeliveredToCompleted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "reported that they have received",
		},
		{name: "Good - Open to Accepted",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromOpenToAccepted,
			sendFunction:     sendNotificationRequestFromOpenToAccepted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "has offered",
		},
		{name: "Bad - Open to Accepted", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromOpenToAccepted,
			sendFunction: sendNotificationRequestFromOpenToAccepted,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestFromOpenToAccepted),
		},
	}

	for _, nextT := range tests {
		t.Run(nextT.name, func(t *testing.T) {

			notifications.TestEmailService.DeleteSentMessages()

			params := senderParams{
				template:   getT(nextT.template),
				subject:    "test subject",
				post:       nextT.post,
				pEventData: nextT.eventData,
			}

			nextT.sendFunction(params)
			gotBuf := buf.String()
			buf.Reset()

			emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
			toEmail := notifications.TestEmailService.GetToEmailByIndex(nextT.wantEmailNumber)
			body := notifications.TestEmailService.GetLastBody()
			ms.Equal(nextT.wantEmailsSent, emailCount, "wrong email count")
			ms.Equal(nextT.wantToEmail, toEmail, "bad To Email")
			ms.Equal(nextT.wantErrLog, gotBuf, "wrong error log entry")
			test.AssertStringContains(t, body, nextT.wantBodyContains, 99)
		})
	}

}

func (ms *ModelSuite) TestSendNewPostNotification() {
	t := ms.T()
	tests := []struct {
		name     string
		user     models.User
		post     models.Post
		wantBody string
		wantErr  string
	}{
		{
			name:    "error - no user email",
			post:    models.Post{UUID: domain.GetUUID(), Title: "post title", Type: models.PostTypeRequest},
			wantErr: "'To' email address is required",
		},
		{
			name: "error - invalid post type",
			user: models.User{
				Email: "user@example.com",
			},
			post:    models.Post{UUID: domain.GetUUID(), Title: "post title", Type: "bogus"},
			wantErr: "invalid template name",
		},
		{
			name: "request",
			user: models.User{
				Email: "user@example.com",
			},
			post:     models.Post{UUID: domain.GetUUID(), Title: "post title", Type: models.PostTypeRequest},
			wantBody: "There is a new request",
		},
		{
			name: "offer",
			user: models.User{
				Email: "user@example.com",
			},
			post:     models.Post{UUID: domain.GetUUID(), Title: "post title", Type: models.PostTypeOffer},
			wantBody: "There is a new offer",
		},
	}
	for _, nextT := range tests {
		t.Run(nextT.name, func(t *testing.T) {
			notifications.TestEmailService.DeleteSentMessages()

			err := sendNewPostNotification(nextT.user, nextT.post)
			if nextT.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), nextT.wantErr)
				return
			}

			ms.NoError(err)

			emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
			ms.Equal(1, emailCount, "wrong email count")

			toEmail := notifications.TestEmailService.GetLastToEmail()
			ms.Equal(nextT.user.Email, toEmail, "bad 'To' address")

			body := notifications.TestEmailService.GetLastBody()

			test.AssertStringContains(t, body, nextT.wantBody, 99)
			test.AssertStringContains(t, body, nextT.post.Title, 99)
			test.AssertStringContains(t, body, nextT.post.UUID.String(), 99)
		})
	}
}

func (ms *ModelSuite) TestSendNewPostNotifications() {
	t := ms.T()
	f := createFixturesForTestSendNewPostNotifications(ms)

	tests := []struct {
		name           string
		post           models.Post
		users          models.Users
		wantEmailCount int
	}{
		{
			name:           "empty",
			post:           f.posts[0],
			wantEmailCount: 0,
		},
		{
			name: "two users",
			post: f.posts[0],
			users: models.Users{
				f.users[1],
				f.users[2],
			},
			wantEmailCount: 2,
		},
		{
			name: "blank in the middle",
			post: f.posts[0],
			users: models.Users{
				f.users[1],
				{Email: ""},
				f.users[2],
			},
			wantEmailCount: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notifications.TestEmailService.DeleteSentMessages()

			sendNewPostNotifications(test.post, test.users)

			emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
			ms.Equal(test.wantEmailCount, emailCount, "wrong email count")

			toAddresses := notifications.TestEmailService.GetAllToAddresses()
			for _, user := range test.users {
				if user.Email == "" {
					continue
				}

				ms.Contains(toAddresses, user.Email, "did not find user address %s", user.Email)
			}
		})
	}
}
