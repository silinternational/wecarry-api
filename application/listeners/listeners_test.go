package listeners

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

// ModelSuite doesn't contain a buffalo suite.Model and can be used for tests that don't need access to the database
// or don't need the buffalo test runner to refresh the database
type ModelSuite struct {
	suite.Suite
	*require.Assertions
	DB *pop.Connection
}

func (ms *ModelSuite) SetupTest() {
	ms.Assertions = require.New(ms.T())
	models.DestroyAll()
}

// Test_ModelSuite runs the test suite
func Test_ModelSuite(t *testing.T) {
	ms := &ModelSuite{}
	c, err := pop.Connect(envy.Get("GO_ENV", "test"))
	if err == nil {
		ms.DB = c
	}
	suite.Run(t, ms)
}

type RequestFixtures struct {
	models.Users
	models.Requests
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

func createFixturesForSendRequestCreatedNotifications(ms *ModelSuite) RequestFixtures {
	users := test.CreateUserFixtures(ms.DB, 3).Users

	request := test.CreateRequestFixtures(ms.DB, 1, false, users[0].ID)[0]
	requestOrigin, err := request.GetOrigin(ms.DB)
	ms.NoError(err)

	for i := range users {
		ms.NoError(users[i].SetLocation(ms.DB, *requestOrigin))
	}

	return RequestFixtures{
		Requests: models.Requests{request},
		Users:    users,
	}
}

func (ms *ModelSuite) TestSendRequestCreatedNotifications() {
	f := createFixturesForSendRequestCreatedNotifications(ms)

	e := events.Event{
		Kind:    domain.EventApiRequestCreated,
		Message: "Request created",
		Payload: events.Payload{domain.EventPayloadKeyEventData: models.RequestCreatedEventData{
			RequestID: f.Requests[0].ID,
		}},
	}

	notifications.TestEmailService.DeleteSentMessages()

	sendRequestCreatedNotifications(e)

	emailsSent := notifications.TestEmailService.GetSentMessages()
	nMessages := 0
	for _, email := range emailsSent {
		if !strings.Contains(email.Subject, "New Request on WeCarry") {
			continue
		}
		if email.ToEmail != f.Users[1].Email && email.ToEmail != f.Users[2].Email {
			continue
		}

		nMessages++
	}
	ms.GreaterOrEqual(nMessages, 2, "wrong email count")
}
