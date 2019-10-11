package listeners

import (
	"bytes"
	"fmt"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/suite"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"
)

type ModelSuite struct {
	*suite.Model
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
	models.ResetTables(ms.T(), ms.DB)

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
