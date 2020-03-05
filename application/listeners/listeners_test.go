package listeners

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/suite"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

type ModelSuite struct {
	*suite.Model
}

type PostFixtures struct {
	models.Users
	models.Posts
}

func Test_ModelSuite(t *testing.T) {
	model := suite.NewModel()

	as := &ModelSuite{
		Model: model,
	}
	suite.Run(t, as)
}

func (ms *ModelSuite) TestRegisterListeners() {
	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	RegisterListeners()
	got := buf.String()

	ms.Equal("", got, "Got an unexpected error log entry")

	wantCount := 0
	for _, listeners := range apiListeners {
		wantCount += len(listeners)
	}

	newLs, err := events.List()
	ms.NoError(err, "Got an unexpected error listing the listeners")

	gotCount := len(newLs)
	ms.Equal(wantCount, gotCount, "Wrong number of listeners registered")

}

// Go through all the listeners that should normally get registered and
// just make sure they don't log anything for a kind they shouldn't be expecting
func (ms *ModelSuite) TestApiListeners_UnusedKind() {
	var buf bytes.Buffer
	domain.Logger.SetOutput(&buf)

	defer func() {
		domain.Logger.SetOutput(os.Stdout)
	}()

	e := events.Event{Kind: "test:unused:kind"}
	for _, listeners := range apiListeners {
		for _, l := range listeners {
			l.listener(e)
			got := buf.String()
			fn := runtime.FuncForPC(reflect.ValueOf(l.listener).Pointer()).Name()
			ms.Equal("", got, fmt.Sprintf("Got an unexpected log entry for listener %s", fn))
		}
	}
}

func (ms *ModelSuite) TestUserCreated() {
	var buf bytes.Buffer
	domain.Logger.SetOutput(&buf)

	defer func() {
		domain.Logger.SetOutput(os.Stdout)
	}()

	user := models.User{
		Email:     "test@test.com",
		FirstName: "test",
		LastName:  "user",
		Nickname:  "testy",
		AdminRole: models.UserAdminRoleUser,
		UUID:      domain.GetUUID(),
	}

	e := events.Event{
		Kind:    domain.EventApiUserCreated,
		Message: "Nickname: " + user.Nickname + "  UUID: " + user.UUID.String(),
		Payload: events.Payload{"user": &user},
	}

	notifications.TestEmailService.DeleteSentMessages()

	userCreatedLogger(e)

	got := buf.String()
	want := fmt.Sprintf("User Created: %s", e.Message)
	test.AssertStringContains(ms.T(), got, want, 74)

	userCreatedSendWelcomeMessage(e)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(1, emailCount, "wrong email count")

}

func (ms *ModelSuite) TestSendNewMessageNotification() {
	var buf bytes.Buffer
	domain.Logger.SetOutput(&buf)

	defer func() {
		domain.Logger.SetOutput(os.Stdout)
	}()

	e := events.Event{
		Kind:    domain.EventApiMessageCreated,
		Message: "New Message from",
	}

	sendNewThreadMessageNotification(e)
	got := buf.String()
	want := "Message Created ... New Message from"

	test.AssertStringContains(ms.T(), got, want, 64)
}

func createFixturesForSendPostCreatedNotifications(ms *ModelSuite) PostFixtures {
	users := test.CreateUserFixtures(ms.DB, 3).Users

	for i := range users {
		ms.NoError(ms.DB.Load(&users[i], "Location"))
		users[i].Location.Country = "KH"
		ms.NoError(ms.DB.Save(&users[i].Location))
	}

	post := test.CreatePostFixtures(ms.DB, 1, false)[0]
	origin := models.Location{Description: "KH", Country: "KH"}
	ms.NoError(post.SetOrigin(origin))

	return PostFixtures{
		Posts: models.Posts{post},
	}
}

func (ms *ModelSuite) TestSendPostCreatedNotifications() {
	f := createFixturesForSendPostCreatedNotifications(ms)

	e := events.Event{
		Kind:    domain.EventApiPostCreated,
		Message: "Post created",
		Payload: events.Payload{"eventData": models.PostCreatedEventData{
			PostID: f.Posts[0].ID,
		}},
	}

	notifications.TestEmailService.DeleteSentMessages()

	sendPostCreatedNotifications(e)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(2, emailCount, "wrong email count")
}
