package listeners

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/silinternational/wecarry-api/notifications"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/suite"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
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

	user := "GoodOne"

	e := events.Event{
		Kind:    domain.EventApiUserCreated,
		Message: user,
	}

	userCreated(e)
	got := buf.String()
	want := "User Created ... " + user

	ms.Contains(got, want, "Got an unexpected log entry")

}

func (ms *ModelSuite) TestUserAccessTokensCleanup() {

	UserAccessTokensNextCleanupTime = time.Now().Add(-time.Duration(time.Hour))

	var buf bytes.Buffer
	domain.Logger.SetOutput(&buf)

	defer func() {
		domain.Logger.SetOutput(os.Stdout)
	}()

	e := events.Event{
		Kind:    domain.EventApiAuthUserLoggedIn,
		Message: "Should get a log",
	}

	userAccessTokensCleanup(e)
	got := buf.String()
	want := "Deleted 0 expired user access tokens during cleanup"

	ms.Contains(got, want, "Got an unexpected log entry")
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

	ms.Contains(got, want, "Got an unexpected log entry")
}

func createFixturesForSendPostCreatedNotifications(ms *ModelSuite) PostFixtures {
	org := models.Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(ms, &org)

	unique := org.Uuid.String()
	users := make(models.Users, 3)
	userLocations := make(models.Locations, len(users))
	userOrgs := make(models.UserOrganizations, len(users))
	for i := range users {
		userLocations[i].Country = "US"
		createFixture(ms, &userLocations[i])

		users[i] = models.User{
			Email:      fmt.Sprintf("%s_user%d@example.com", unique, i),
			Nickname:   fmt.Sprintf("%s_User%d", unique, i),
			Uuid:       domain.GetUuid(),
			LocationID: nulls.NewInt(userLocations[i].ID),
		}
		createFixture(ms, &users[i])

		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].UserID = users[i].ID
		userOrgs[i].AuthEmail = users[i].Email
		userOrgs[i].AuthID = users[i].Email
		createFixture(ms, &(userOrgs[i]))
	}

	location := models.Location{Country: "US"}
	createFixture(ms, &location)

	post := models.Post{
		OrganizationID: org.ID,
		Uuid:           domain.GetUuid(),
		CreatedByID:    users[0].ID,
		DestinationID:  location.ID,
		Type:           models.PostTypeOffer,
	}
	createFixture(ms, &post)

	return PostFixtures{
		Users: users,
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
