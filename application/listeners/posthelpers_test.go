package listeners

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

func (ms *ModelSuite) TestGetPostUsers() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_GetPostRecipients(ms, t)
	users := orgUserPostFixtures.users
	posts := orgUserPostFixtures.posts

	tests := []struct {
		name          string
		id            int
		wantRequester PostUser
		wantProvider  PostUser
		wantErr       bool
	}{
		{name: "Request by User0 with User1 as Provider",
			id: posts[0].ID,
			wantRequester: PostUser{
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
			wantProvider: PostUser{
				Nickname: users[1].Nickname,
				Email:    users[1].Email,
			},
		},
		{name: "Request by User0 with no Provider",
			id: posts[1].ID,
			wantRequester: PostUser{
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

			postUsers := GetPostUsers(post)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantRequester, postUsers.Requester)
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
		NewStatus: models.PostStatusCommitted,
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
	ms.Contains(got, "unexpected status transition 'OPEN-ACCEPTED'", "Got an unexpected error log entry")

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
		{name: "Good - Accepted to Committed",
			post: posts[0],
			eventData: models.PostStatusEventData{
				OldStatus:     models.PostStatusAccepted,
				NewStatus:     models.PostStatusCommitted,
				PostID:        posts[0].ID,
				OldProviderID: *models.GetIntFromNullsInt(posts[0].ProviderID),
			},
			template:         domain.MessageTemplateRequestFromAcceptedToCommitted,
			sendFunction:     sendNotificationRequestFromAcceptedToCommitted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "isn't sure yet if they",
		},
		{name: "Good - Accepted to Completed",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromAcceptedToCompleted,
			sendFunction:     sendNotificationRequestFromAcceptedOrDeliveredToCompleted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "Thank you for fulfilling a request",
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
			wantBodyContains: "says they have delivered it",
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
			wantBodyContains: "no longer wants you",
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
		{name: "Good - Committed to Accepted",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromCommittedToAccepted,
			sendFunction:     sendNotificationRequestFromCommittedToAccepted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "has accepted your offer",
		},
		{name: "Bad - Committed to Accepted", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromCommittedToAccepted,
			sendFunction: sendNotificationRequestFromCommittedToAccepted,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestFromCommittedToAccepted),
		},
		{name: "Good - Committed to Delivered",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromCommittedToDelivered,
			sendFunction:     sendNotificationRequestFromAcceptedOrCommittedToDelivered,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "they have delivered it",
		},
		{name: "Bad - Committed to Delivered", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromCommittedToDelivered,
			sendFunction: sendNotificationRequestFromAcceptedOrCommittedToDelivered,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				getT(domain.MessageTemplateRequestFromCommittedToDelivered)),
		},
		{name: "Good - Committed to Open - Provider",
			post: posts[0],
			eventData: models.PostStatusEventData{
				OldStatus:     models.PostStatusCommitted,
				NewStatus:     models.PostStatusOpen,
				PostID:        posts[0].ID,
				OldProviderID: *models.GetIntFromNullsInt(posts[0].ProviderID),
			},
			template:         domain.MessageTemplateRequestFromCommittedToOpen,
			sendFunction:     sendNotificationRequestFromCommittedToOpen,
			wantEmailsSent:   2,
			wantEmailNumber:  1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "no longer has a provider",
		},
		{name: "Good - Committed to Open - Requester",
			post: posts[0],
			eventData: models.PostStatusEventData{
				OldStatus:     models.PostStatusCommitted,
				NewStatus:     models.PostStatusOpen,
				PostID:        posts[0].ID,
				OldProviderID: *models.GetIntFromNullsInt(posts[0].ProviderID),
			},
			template:         domain.MessageTemplateRequestFromCommittedToOpen,
			sendFunction:     sendNotificationRequestFromCommittedToOpen,
			wantEmailsSent:   2,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "no longer has a provider",
		},
		{name: "Good - Committed to Removed",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromCommittedToRemoved,
			sendFunction:     sendNotificationRequestFromCommittedToRemoved,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "has removed this request",
		},
		{name: "Bad - Committed to Removed", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromCommittedToRemoved,
			sendFunction: sendNotificationRequestFromCommittedToRemoved,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestFromCommittedToRemoved),
		},
		{name: "Good - Completed to Accepted",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromCompletedToAccepted,
			sendFunction:     sendNotificationRequestFromCompletedToAcceptedOrDelivered,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "said they haven't received the item afterall",
		},
		{name: "Bad - Completed to Accepted", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromCompletedToAccepted,
			sendFunction: sendNotificationRequestFromCompletedToAcceptedOrDelivered,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestNotReceivedAfterall),
		},
		{name: "Good - Delivered to Accepted",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromDeliveredToAccepted,
			sendFunction:     sendNotificationRequestFromDeliveredToAccepted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "now says they haven't yet delivered it",
		},
		{name: "Good - Delivered to Committed",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromDeliveredToCommitted,
			sendFunction:     sendNotificationRequestFromDeliveredToCommitted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "now says they haven't yet delivered it",
		},
		{name: "Good - Delivered to Completed",
			post:             posts[0],
			template:         domain.MessageTemplateRequestReceived,
			sendFunction:     sendNotificationRequestFromAcceptedOrDeliveredToCompleted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].Provider.Email,
			wantBodyContains: "Thank you for fulfilling a request",
		},
		{name: "Good - Open to Committed",
			post:             posts[0],
			template:         domain.MessageTemplateRequestFromOpenToCommitted,
			sendFunction:     sendNotificationRequestFromOpenToCommitted,
			wantEmailsSent:   1,
			wantToEmail:      posts[0].CreatedBy.Email,
			wantBodyContains: "has offered",
		},
		{name: "Bad - Open to Committed", // No Provider
			post:         posts[1],
			template:     domain.MessageTemplateRequestFromOpenToCommitted,
			sendFunction: sendNotificationRequestFromOpenToCommitted,
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestFromOpenToCommitted),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			notifications.TestEmailService.DeleteSentMessages()

			params := senderParams{
				template:   getT(test.template),
				subject:    "test subject",
				post:       test.post,
				pEventData: test.eventData,
			}

			test.sendFunction(params)
			gotBuf := buf.String()
			buf.Reset()

			emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
			toEmail := notifications.TestEmailService.GetToEmailByIndex(test.wantEmailNumber)
			body := notifications.TestEmailService.GetLastBody()
			ms.Equal(test.wantEmailsSent, emailCount, "wrong email count")
			ms.Equal(test.wantToEmail, toEmail, "bad To Email")
			ms.Equal(test.wantErrLog, gotBuf, "wrong error log entry")
			ms.Contains(body, test.wantBodyContains, "Body doesn't contain expected string")
		})
	}

}

func (ms *ModelSuite) TestGetTranslatedSubject() {
	t := ms.T()

	tests := []struct {
		name          string
		translationID string
		want          string
	}{
		{
			name:          "delivered",
			translationID: "Email.Subject.Request.FromAcceptedOrCommittedToDelivered",
			want:          "Request marked as delivered on " + domain.Env.AppName,
		},
		{
			name:          "from accepted to committed",
			translationID: "Email.Subject.Request.FromAcceptedToCommitted",
			want:          "Oops, you are not yet expected to fulfill a certain " + domain.Env.AppName + " request",
		},
		{
			name:          "from accepted to completed",
			translationID: "Email.Subject.Request.FromAcceptedOrDeliveredToCompleted",
			want:          "Thank you for fulfilling a request on " + domain.Env.AppName,
		},
		{
			name:          "from accepted to open",
			translationID: "Email.Subject.Request.FromAcceptedToOpen",
			want:          "You are no longer expected to fulfill a certain " + domain.Env.AppName + " request",
		},
		{
			name:          "from accepted to removed",
			translationID: "Email.Subject.Request.FromAcceptedToRemoved",
			want:          "You are no longer expected to fulfill a certain " + domain.Env.AppName + " request",
		},
		{
			name:          "from committed to accepted",
			translationID: "Email.Subject.Request.FromCommittedToAccepted",
			want:          "Your offer was accepted on " + domain.Env.AppName,
		},
		{
			name:          "from committed to open",
			translationID: "Email.Subject.Request.FromCommittedToOpen",
			want:          "Request lost its provider on " + domain.Env.AppName,
		},
		{
			name:          "from committed to removed",
			translationID: "Email.Subject.Request.FromCommittedToRemoved",
			want:          "Request removed on " + domain.Env.AppName,
		},
		{
			name:          "from completed to accepted",
			translationID: "Email.Subject.Request.FromCompletedToAcceptedOrDelivered",
			want:          "Oops, request not received on " + domain.Env.AppName + " afterall",
		},
		{
			name:          "from delivered to accepted",
			translationID: "Email.Subject.Request.FromDeliveredToAccepted",
			want:          "Request not delivered afterall on " + domain.Env.AppName,
		},
		{
			name:          "from delivered to committed",
			translationID: "Email.Subject.Request.FromDeliveredToCommitted",
			want:          "Request not delivered afterall on " + domain.Env.AppName,
		},
		{
			name:          "from open to committed",
			translationID: "Email.Subject.Request.FromOpenToCommitted",
			want:          "Potential provider on " + domain.Env.AppName,
		},
	}

	template := "test subject"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getTranslatedSubject(test.translationID, template)
			ms.Equal(test.want, got, "bad subject translation")
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
			post:    models.Post{Uuid: domain.GetUuid(), Title: "post title", Type: models.PostTypeRequest},
			wantErr: "'To' email address is required",
		},
		{
			name: "error - invalid post type",
			user: models.User{
				Email: "user@example.com",
			},
			post:    models.Post{Uuid: domain.GetUuid(), Title: "post title", Type: "bogus"},
			wantErr: "invalid template name",
		},
		{
			name: "request",
			user: models.User{
				Email: "user@example.com",
			},
			post:     models.Post{Uuid: domain.GetUuid(), Title: "post title", Type: models.PostTypeRequest},
			wantBody: "There is a new request",
		},
		{
			name: "offer",
			user: models.User{
				Email: "user@example.com",
			},
			post:     models.Post{Uuid: domain.GetUuid(), Title: "post title", Type: models.PostTypeOffer},
			wantBody: "There is a new offer",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notifications.TestEmailService.DeleteSentMessages()

			err := sendNewPostNotification(test.user, test.post)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}

			ms.NoError(err)

			emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
			ms.Equal(1, emailCount, "wrong email count")

			toEmail := notifications.TestEmailService.GetLastToEmail()
			ms.Equal(test.user.Email, toEmail, "bad 'To' address")

			body := notifications.TestEmailService.GetLastBody()
			ms.Contains(body, test.wantBody, "Body doesn't contain expected string")
			ms.Contains(body, test.post.Title, "Body doesn't contain post title")
			ms.Contains(body, test.post.Uuid.String(), "Body doesn't contain post UUID")
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
