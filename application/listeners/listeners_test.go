package listeners

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/log"
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
	log.SetOutput(&buf)

	defer func() {
		log.SetOutput(os.Stderr)
	}()

	RegisterListener()
	got := buf.String()

	ms.Equal("", got, "Got an unexpected error log entry")

	newLs, err := events.List()
	ms.NoError(err, "Got an unexpected error listing the listeners")

	gotCount := len(newLs)
	ms.Equal(1, gotCount, "Wrong number of listeners registered")
}

func (ms *ModelSuite) TestUserCreated() {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	defer func() {
		log.SetOutput(os.Stdout)
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

	// got := buf.String()
	// want := fmt.Sprintf("User Created: %s", e.Message)
	// test.AssertStringContains(ms.T(), got, want, 74)

	userCreatedSendWelcomeMessage(e)

	emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
	ms.Equal(1, emailCount, "wrong email count")
}

func (ms *ModelSuite) TestSendNewMessageNotification() {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	defer func() {
		log.SetOutput(os.Stdout)
	}()

	e := events.Event{
		Kind:    domain.EventApiMessageCreated,
		Message: "New Message from",
	}

	sendNewThreadMessageNotification(e)
	// got := buf.String()
	// want := "Message Created ... New Message from"

	// test.AssertStringContains(ms.T(), got, want, 64)
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
		Payload: events.Payload{domain.ArgEventData: models.RequestCreatedEventData{
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

func (ms *ModelSuite) TestMeetingInviteCreated() {
	users := test.CreateUserFixtures(ms.DB, 3).Users

	meeting := test.CreateMeetingFixtures(ms.DB, 1, users[0])[0]

	notifications.TestEmailService.DeleteSentMessages()

	ms.NoError(meeting.CreateInvites(test.CtxWithUser(users[0]), "test@example.com"))

	invite := models.MeetingInvite{}
	ms.NoError(ms.DB.Where("meeting_id = ?", meeting.ID).First(&invite))

	// in test, there is no listener goroutine, so we have to fake it and call the function directly
	meetingInviteCreated(events.Event{
		Kind:    domain.EventApiMeetingInviteCreated,
		Message: "Meeting Invite created",
		Payload: events.Payload{domain.ArgId: invite.ID},
	})

	emailsSent := notifications.TestEmailService.GetSentMessages()

	nMessages := 0
	for _, email := range emailsSent {
		if !strings.Contains(email.Subject, "Invitation to "+meeting.Name) {
			continue
		}

		if email.ToEmail != invite.Email {
			continue
		}

		nMessages++
	}
	ms.Equal(nMessages, 1, "wrong email count")
}
