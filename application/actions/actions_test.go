package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/pop/v5"
	"github.com/gorilla/sessions"
	"github.com/silinternational/wecarry-api/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ActionSuite struct {
	suite.Suite
	*require.Assertions
	App     *buffalo.App
	DB      *pop.Connection
	Session *buffalo.Session
}

// HTML creates an httptest.Request with HTML content type.
func (as *ActionSuite) HTML(u string, args ...interface{}) *httptest.Request {
	return httptest.New(as.App).HTML(u, args...)
}

// JSON creates an httptest.JSON request
func (as *ActionSuite) JSON(u string, args ...interface{}) *httptest.JSON {
	return httptest.New(as.App).JSON(u, args...)
}

func Test_ActionSuite(t *testing.T) {
	as := &ActionSuite{
		App: App(),
	}
	c, err := pop.Connect("test")
	if err == nil {
		as.DB = c
	}
	suite.Run(t, as)
}

// SetupTest sets the test suite to abort on first failure and sets the session store
func (as *ActionSuite) SetupTest() {
	as.Assertions = require.New(as.T())

	as.App.SessionStore = newSessionStore()
	s, _ := as.App.SessionStore.New(nil, as.App.SessionName)
	as.Session = &buffalo.Session{
		Session: s,
	}

	models.DestroyAll()
}

// sessionStore copied from gobuffalo/suite session.go
type sessionStore struct {
	sessions map[string]*sessions.Session
}

func (s *sessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	if s, ok := s.sessions[name]; ok {
		return s, nil
	}
	return s.New(r, name)
}

func (s *sessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	sess := sessions.NewSession(s, name)
	s.sessions[name] = sess
	return sess, nil
}

func (s *sessionStore) Save(r *http.Request, w http.ResponseWriter, sess *sessions.Session) error {
	if s.sessions == nil {
		s.sessions = map[string]*sessions.Session{}
	}
	s.sessions[sess.Name()] = sess
	return nil
}

// NewSessionStore for action suite
func newSessionStore() sessions.Store {
	return &sessionStore{
		sessions: map[string]*sessions.Session{},
	}
}

func createFixture(as *ActionSuite, f interface{}) {
	err := as.DB.Create(f)
	if err != nil {
		as.T().Errorf("error creating %T fixture, %s", f, err)
		as.T().FailNow()
	}
}

// Avoid issues with int(-0.xyz) losing its negative sign
func convertFloat64ToIntString(input float64) string {
	if -1.0 < input && input < 0.0 {
		return fmt.Sprintf("-%v", int(input))
	}
	return fmt.Sprintf("%v", int(input))
}

func (as *ActionSuite) verifyResponseData(wantData []string, body string) {
	var b bytes.Buffer
	as.NoError(json.Indent(&b, []byte(body), "", "    "))
	for _, w := range wantData {
		if !strings.Contains(body, w) {
			as.Fail(fmt.Sprintf("response data is not correct\nwanted: %s\nin body:\n%s\n", w, b.String()))
		}
	}
}
