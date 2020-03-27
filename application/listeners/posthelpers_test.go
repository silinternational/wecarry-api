package listeners

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gobuffalo/nulls"

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

			var post models.Request
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

	postStatusEData := models.RequestStatusEventData{
		OldStatus: models.RequestStatusOpen,
		NewStatus: models.RequestStatusAccepted,
		RequestID: posts[0].ID,
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
	postStatusEData.NewStatus = models.RequestStatusDelivered
	requestStatusUpdatedNotifications(posts[0], postStatusEData)

	got = buf.String()
	want := "unexpected status transition 'OPEN-DELIVERED'"
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
		post             models.Request
		eventData        models.RequestStatusEventData
		sendFunction     func(senderParams)
		wantEmailsSent   int
		wantToEmail      string
		wantBodyContains string
		wantEmailNumber  int
		wantErrLog       string
	}{
		{name: "Good - Accepted to Open",
			post: posts[0],
			eventData: models.RequestStatusEventData{
				OldStatus:     models.RequestStatusAccepted,
				NewStatus:     models.RequestStatusOpen,
				RequestID:     posts[0].ID,
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
			sendFunction:     sendNotificationRequestFromAcceptedToDelivered,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "reported that they have delivered your request",
		},
		{name: "Bad - Accepted to Delivered", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromAcceptedToDelivered,
			sendFunction: sendNotificationRequestFromAcceptedToDelivered,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				getT(domain.MessageTemplateRequestFromAcceptedToDelivered)),
		},
		{name: "Good - Accepted to Open",
			post: posts[0],
			eventData: models.RequestStatusEventData{
				OldStatus:     models.RequestStatusAccepted,
				NewStatus:     models.RequestStatusOpen,
				RequestID:     posts[0].ID,
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
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "has accepted your offer",
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
	post := test.CreateRequestFixtures(ms.DB, 1, false)[0]
	tests := []struct {
		name     string
		user     models.User
		post     models.Request
		wantBody string
		wantErr  string
	}{
		{
			name:    "error - no user email",
			post:    post,
			wantErr: "'To' email address is required",
		},
		{
			name: "request",
			user: models.User{
				Email: "user@example.com",
			},
			post:     post,
			wantBody: "There is a new request",
		},
	}
	for _, nextT := range tests {
		ms.T().Run(nextT.name, func(t *testing.T) {
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
		post           models.Request
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

func (ms *ModelSuite) TestSendPotentialProviderCreatedNotification() {
	t := ms.T()
	provider := "Pete Provider"
	requester := models.User{
		Email: "user@example.com",
	}
	post := models.Request{UUID: domain.GetUUID(), Title: "post title"}
	wantBody := "has offered to help fulfill your request"

	notifications.TestEmailService.DeleteSentMessages()

	err := sendPotentialProviderCreatedNotification(provider, requester, post)
	ms.NoError(err)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(1, emailCount, "wrong email count")

	toEmail := notifications.TestEmailService.GetLastToEmail()
	ms.Equal(requester.Email, toEmail, "bad 'To' address")

	body := notifications.TestEmailService.GetLastBody()

	test.AssertStringContains(t, body, wantBody, 99)
	test.AssertStringContains(t, body, post.Title, 99)
	test.AssertStringContains(t, body, post.UUID.String(), 99)
}

func (ms *ModelSuite) TestSendPotentialProviderSelfDestroyedNotification() {
	t := ms.T()
	provider := "Pete Provider"
	requester := models.User{
		Email: "user@example.com",
	}
	post := models.Request{UUID: domain.GetUUID(), Title: "post title"}
	wantBody := "indicated they can't fulfill your request afterall"

	notifications.TestEmailService.DeleteSentMessages()

	err := sendPotentialProviderSelfDestroyedNotification(provider, requester, post)
	ms.NoError(err)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(1, emailCount, "wrong email count")

	toEmail := notifications.TestEmailService.GetLastToEmail()
	ms.Equal(requester.Email, toEmail, "bad 'To' address")

	body := notifications.TestEmailService.GetLastBody()

	test.AssertStringContains(t, body, wantBody, 99)
	test.AssertStringContains(t, body, post.Title, 99)
	test.AssertStringContains(t, body, post.UUID.String(), 99)
}

func (ms *ModelSuite) TestSendPotentialProviderRejectedNotification() {
	t := ms.T()
	requester := "Rodger Requester"
	provider := models.User{
		Email: "user@example.com",
	}
	post := models.Request{UUID: domain.GetUUID(), Title: "post title"}
	wantBody := "is not prepared to have you fulfill their request"

	notifications.TestEmailService.DeleteSentMessages()

	err := sendPotentialProviderRejectedNotification(provider, requester, post)
	ms.NoError(err)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(1, emailCount, "wrong email count")

	toEmail := notifications.TestEmailService.GetLastToEmail()
	ms.Equal(provider.Email, toEmail, "bad 'To' address")

	body := notifications.TestEmailService.GetLastBody()

	test.AssertStringContains(t, body, wantBody, 99)
	test.AssertStringContains(t, body, post.Title, 99)
	test.AssertStringContains(t, body, post.UUID.String(), 99)
}

func (ms *ModelSuite) TestSendNotificationRequestFromOpenToAccepted() {
	// Five User and three Request fixtures will also be created.  The Requests will
	// all be created by the first user.
	// The first Request will have all but the first and fifth user as a potential provider.
	f := test.CreatePotentialProvidersFixtures(ms.DB)

	users := f.Users
	post := f.Requests[0]

	post.ProviderID = nulls.NewInt(f.Users[3].ID)

	notifications.TestEmailService.DeleteSentMessages()

	eData := models.RequestStatusEventData{
		OldStatus: models.RequestStatusOpen,
		NewStatus: models.RequestStatusAccepted,
		RequestID: post.ID,
	}

	params := senderParams{
		template:   domain.MessageTemplateRequestFromOpenToAccepted,
		subject:    "Email.Subject.Request.FromOpenToAccepted",
		post:       post,
		pEventData: eData,
	}

	sendNotificationRequestFromOpenToAccepted(params)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.GreaterOrEqual(emailCount, 3, "wrong email count")

	// Fourth user is Provider, so he should get acceptance email and
	// all the others should get the rejection email
	sentMsgs := notifications.TestEmailService.GetSentMessages()

	// Timing issues are allowing older emails from the fixture creation to leak
	// into the sentEmails after the next DeletSentMessages() call
	acceptedSubject := "request has been accepted"
	rejectedSubject := "was not accepted"

	accepteds := []notifications.DummyMessageInfo{}
	rejects := []notifications.DummyMessageInfo{}

	for _, m := range sentMsgs {
		if strings.Contains(m.Subject, acceptedSubject) {
			accepteds = append(accepteds, m)
			continue
		}

		if strings.Contains(m.Subject, rejectedSubject) {
			rejects = append(rejects, m)
			continue
		}
	}

	ms.Equal(1, len(accepteds), "incorrect number of accepted messages")
	ms.Equal(users[3].Email, accepteds[0].ToEmail, "incorrect recipient for accepted message")

	ms.Equal(2, len(rejects), "incorrect number of rejected messages")

}
