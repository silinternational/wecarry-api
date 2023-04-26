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
	"github.com/silinternational/wecarry-api/log"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

func (ms *ModelSuite) TestGetRequestUsers() {
	t := ms.T()

	orgUserRequestFixtures := CreateFixtures_GetRequestUsers(ms, t)
	users := orgUserRequestFixtures.users
	requests := orgUserRequestFixtures.requests

	tests := []struct {
		name          string
		id            int
		wantRequester requestUser
		wantProvider  requestUser
		wantErr       bool
	}{
		{
			name: "Request by User0 with User1 as Provider",
			id:   requests[0].ID,
			wantRequester: requestUser{
				Language: domain.UserPreferenceLanguageEnglish,
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
			wantProvider: requestUser{
				Language: domain.UserPreferenceLanguageFrench,
				Nickname: users[1].Nickname,
				Email:    users[1].Email,
			},
		},
		{
			name: "Request by User0 with no Provider",
			id:   requests[1].ID,
			wantRequester: requestUser{
				Language: domain.UserPreferenceLanguageEnglish,
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var request models.Request
			err := request.FindByID(ms.DB, test.id)
			ms.NoError(err, "error finding request for test")

			requestUsers := getRequestUsers(request)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantRequester, requestUsers.Receiver)
				ms.Equal(test.wantProvider, requestUsers.Provider)
			}
		})
	}
}

func (ms *ModelSuite) TestRequestStatusUpdatedNotifications() {
	t := ms.T()

	orgUserRequestFixtures := CreateFixtures_RequestStatusUpdatedNotifications(ms, t)
	requests := orgUserRequestFixtures.requests

	requestStatusEData := models.RequestStatusEventData{
		OldStatus: models.RequestStatusOpen,
		NewStatus: models.RequestStatusAccepted,
		RequestID: requests[0].ID,
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)

	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// No logging message expected
	requestStatusUpdatedNotifications(requests[0], requestStatusEData)

	got := buf.String()
	ms.Equal("", got, "Got an unexpected error log entry")

	buf.Reset()

	// Logging message expected about bad transition
	requestStatusEData.NewStatus = models.RequestStatusDelivered
	requestStatusUpdatedNotifications(requests[0], requestStatusEData)

	// got = buf.String()
	// want := "unexpected request status transition 'OPEN-DELIVERED'"
	// test.AssertStringContains(t, got, want, 45)
}

func (ms *ModelSuite) TestSendNotificationRequestFromStatus() {
	t := ms.T()

	orgUserRequestFixtures := CreateFixtures_sendNotificationRequestFromStatus(ms, t)
	requests := orgUserRequestFixtures.requests

	var buf bytes.Buffer
	log.SetOutput(&buf)

	defer func() {
		log.SetOutput(os.Stderr)
	}()

	getT := notifications.GetEmailTemplate

	tests := []struct {
		name             string
		template         string
		request          models.Request
		eventData        models.RequestStatusEventData
		sendFunction     func(senderParams)
		wantEmailsSent   int
		wantToEmail      string
		wantBodyContains string
		wantEmailNumber  int
		wantErrLog       string
	}{
		{
			name:    "Good - Accepted to Open",
			request: requests[0],
			eventData: models.RequestStatusEventData{
				OldStatus:     models.RequestStatusAccepted,
				NewStatus:     models.RequestStatusOpen,
				RequestID:     requests[0].ID,
				OldProviderID: *models.GetIntFromNullsInt(requests[0].ProviderID),
			},
			template:         domain.MessageTemplateRequestFromAcceptedToOpen,
			sendFunction:     sendNotificationRequestFromAcceptedToOpen,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].Provider.Email,
			wantBodyContains: "isn't ready after all to have you fulfill",
		},
		{
			name:             "Good - Accepted to Completed",
			request:          requests[0],
			template:         domain.MessageTemplateRequestFromAcceptedToCompleted,
			sendFunction:     sendNotificationRequestFromAcceptedOrDeliveredToCompleted,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].Provider.Email,
			wantBodyContains: "reported that they have received",
		},
		{
			name:         "Bad - Accepted to Completed", // No Provider
			request:      requests[1],
			template:     domain.MessageTemplateRequestFromAcceptedToCompleted,
			sendFunction: sendNotificationRequestFromAcceptedOrDeliveredToCompleted,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestReceived),
		},
		{
			name:             "Good - Accepted to Delivered",
			request:          requests[0],
			template:         domain.MessageTemplateRequestDelivered,
			sendFunction:     sendNotificationRequestFromAcceptedToDelivered,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].CreatedBy.Email,
			wantBodyContains: "reported that they have delivered your request",
		},
		{
			name:         "Bad - Accepted to Delivered", // No Provider
			request:      requests[1],
			template:     domain.MessageTemplateRequestFromAcceptedToDelivered,
			sendFunction: sendNotificationRequestFromAcceptedToDelivered,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				getT(domain.MessageTemplateRequestFromAcceptedToDelivered)),
		},
		{
			name:    "Good - Accepted to Open",
			request: requests[0],
			eventData: models.RequestStatusEventData{
				OldStatus:     models.RequestStatusAccepted,
				NewStatus:     models.RequestStatusOpen,
				RequestID:     requests[0].ID,
				OldProviderID: *models.GetIntFromNullsInt(requests[0].ProviderID),
			},
			template:         domain.MessageTemplateRequestFromAcceptedToOpen,
			sendFunction:     sendNotificationRequestFromAcceptedToOpen,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].Provider.Email,
			wantBodyContains: "isn't ready after all to have you fulfill",
		},
		{
			name:             "Good - Accepted to Removed",
			request:          requests[0],
			template:         domain.MessageTemplateRequestFromAcceptedToRemoved,
			sendFunction:     sendNotificationRequestFromAcceptedToRemoved,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].Provider.Email,
			wantBodyContains: "has removed this request",
		},
		{
			name:         "Bad - Accepted to Removed", // No Provider
			request:      requests[1],
			template:     domain.MessageTemplateRequestFromAcceptedToRemoved,
			sendFunction: sendNotificationRequestFromAcceptedToRemoved,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestFromAcceptedToRemoved),
		},
		{
			name:             "Good - Completed to Accepted",
			request:          requests[0],
			template:         domain.MessageTemplateRequestFromCompletedToAccepted,
			sendFunction:     sendNotificationRequestFromCompletedToAcceptedOrDelivered,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].Provider.Email,
			wantBodyContains: "but now have corrected that",
		},
		{
			name:         "Bad - Completed to Accepted", // No Provider
			request:      requests[1],
			template:     domain.MessageTemplateRequestFromCompletedToAccepted,
			sendFunction: sendNotificationRequestFromCompletedToAcceptedOrDelivered,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestNotReceivedAfterAll),
		},
		{
			name:             "Good - Delivered to Accepted",
			request:          requests[0],
			template:         domain.MessageTemplateRequestFromDeliveredToAccepted,
			sendFunction:     sendNotificationRequestFromDeliveredToAccepted,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].CreatedBy.Email,
			wantBodyContains: "corrected themselves to say they haven't",
		},
		{
			name:             "Good - Delivered to Completed",
			request:          requests[0],
			template:         domain.MessageTemplateRequestReceived,
			sendFunction:     sendNotificationRequestFromAcceptedOrDeliveredToCompleted,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].Provider.Email,
			wantBodyContains: "reported that they have received",
		},
		{
			name:             "Good - Open to Accepted",
			request:          requests[0],
			template:         domain.MessageTemplateRequestFromOpenToAccepted,
			sendFunction:     sendNotificationRequestFromOpenToAccepted,
			wantEmailsSent:   1,
			wantToEmail:      requests[0].Provider.Email,
			wantBodyContains: "has accepted your offer",
		},
		{
			name:         "Bad - Open to Accepted", // No Provider
			request:      requests[1],
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
				request:    nextT.request,
				pEventData: nextT.eventData,
			}

			nextT.sendFunction(params)
			// gotBuf := buf.String()
			buf.Reset()

			emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
			toEmail := notifications.TestEmailService.GetToEmailByIndex(nextT.wantEmailNumber)
			body := notifications.TestEmailService.GetLastBody()
			ms.Equal(nextT.wantEmailsSent, emailCount, "wrong email count")
			ms.Equal(nextT.wantToEmail, toEmail, "bad To Email")
			// ms.Equal(nextT.wantErrLog, gotBuf, "wrong error log entry")
			test.AssertStringContains(t, body, nextT.wantBodyContains, 99)
		})
	}
}

func (ms *ModelSuite) TestSendNewRequestNotification() {
	request := test.CreateRequestFixtures(ms.DB, 1, false)[0]
	tests := []struct {
		name     string
		user     models.User
		request  models.Request
		wantBody string
		wantErr  string
	}{
		{
			name:    "error - no user email",
			request: request,
			wantErr: "'To' email address is required",
		},
		{
			name: "request",
			user: models.User{
				Email: "user@example.com",
			},
			request:  request,
			wantBody: "There is a new request",
		},
	}
	for _, nextT := range tests {
		ms.T().Run(nextT.name, func(t *testing.T) {
			notifications.TestEmailService.DeleteSentMessages()

			err := sendNewRequestNotification(nextT.user, nextT.request)
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
			test.AssertStringContains(t, body, nextT.request.Title, 99)
			test.AssertStringContains(t, body, nextT.request.UUID.String(), 99)
		})
	}
}

func (ms *ModelSuite) TestSendNewRequestNotifications() {
	t := ms.T()
	f := createFixturesForTestSendNewRequestNotifications(ms)

	tests := []struct {
		name           string
		request        models.Request
		users          models.Users
		wantEmailCount int
	}{
		{
			name:           "empty",
			request:        f.requests[0],
			wantEmailCount: 0,
		},
		{
			name:    "two users",
			request: f.requests[0],
			users: models.Users{
				f.users[1],
				f.users[2],
			},
			wantEmailCount: 2,
		},
		{
			name:    "blank in the middle",
			request: f.requests[0],
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

			sendNewRequestNotifications(test.request, test.users)

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
	request := models.Request{UUID: domain.GetUUID(), Title: "request title"}
	wantBody := "has offered to help fulfill your request"

	notifications.TestEmailService.DeleteSentMessages()

	err := sendPotentialProviderCreatedNotification(provider, requester, request)
	ms.NoError(err)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(1, emailCount, "wrong email count")

	toEmail := notifications.TestEmailService.GetLastToEmail()
	ms.Equal(requester.Email, toEmail, "bad 'To' address")

	body := notifications.TestEmailService.GetLastBody()

	test.AssertStringContains(t, body, wantBody, 99)
	test.AssertStringContains(t, body, request.Title, 99)
	test.AssertStringContains(t, body, request.UUID.String(), 99)
}

func (ms *ModelSuite) TestSendPotentialProviderSelfDestroyedNotification() {
	t := ms.T()
	provider := "Pete Provider"
	requester := models.User{
		Email: "user@example.com",
	}
	request := models.Request{UUID: domain.GetUUID(), Title: "request title"}
	wantBody := "indicated they can't fulfill your request afterall"

	notifications.TestEmailService.DeleteSentMessages()

	err := sendPotentialProviderSelfDestroyedNotification(provider, requester, request)
	ms.NoError(err)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(1, emailCount, "wrong email count")

	toEmail := notifications.TestEmailService.GetLastToEmail()
	ms.Equal(requester.Email, toEmail, "bad 'To' address")

	body := notifications.TestEmailService.GetLastBody()

	test.AssertStringContains(t, body, wantBody, 99)
	test.AssertStringContains(t, body, request.Title, 99)
	test.AssertStringContains(t, body, request.UUID.String(), 99)
}

func (ms *ModelSuite) TestSendPotentialProviderRejectedNotification() {
	t := ms.T()
	requester := "Rodger Requester"
	provider := models.User{
		Email: "user@example.com",
	}
	request := models.Request{UUID: domain.GetUUID(), Title: "request title"}
	wantBody := "is not prepared to have you fulfill their request"

	notifications.TestEmailService.DeleteSentMessages()

	err := sendPotentialProviderRejectedNotification(provider, requester, request)
	ms.NoError(err)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(1, emailCount, "wrong email count")

	toEmail := notifications.TestEmailService.GetLastToEmail()
	ms.Equal(provider.Email, toEmail, "bad 'To' address")

	body := notifications.TestEmailService.GetLastBody()

	test.AssertStringContains(t, body, wantBody, 99)
	test.AssertStringContains(t, body, request.Title, 99)
	test.AssertStringContains(t, body, request.UUID.String(), 99)
}

func (ms *ModelSuite) TestSendNotificationRequestFromOpenToAccepted() {
	// Five User and three Request fixtures will also be created.  The Requests will
	// all be created by the first user.
	// The first Request will have all but the first and fifth user as a potential provider.
	f := test.CreatePotentialProvidersFixtures(ms.DB)

	users := f.Users
	request := f.Requests[0]

	request.ProviderID = nulls.NewInt(f.Users[3].ID)

	notifications.TestEmailService.DeleteSentMessages()

	eData := models.RequestStatusEventData{
		OldStatus: models.RequestStatusOpen,
		NewStatus: models.RequestStatusAccepted,
		RequestID: request.ID,
	}

	params := senderParams{
		template:   domain.MessageTemplateRequestFromOpenToAccepted,
		subject:    "Email.Subject.Request.FromOpenToAccepted",
		request:    request,
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
