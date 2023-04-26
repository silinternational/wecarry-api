package aws

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
)

// TestSuite establishes a test suite
type TestSuite struct {
	suite.Suite
}

// Test_TestSuite runs the test suite
func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// TestSendEmail can be used in a local environment for development. Add SES credentials to the appropriate
// environment variables, and change the "To" and "From" email addresses to valid addresses.
func (ts *TestSuite) TestSendEmail() {
	ts.T().Skip("only for use in local environment if configured with SES credentials")
	err := SendEmail(
		"me@example.com",
		domain.Env.EmailFromAddress,
		"test subject",
		`<h4>body</h4><img src="cid:logo"><p>End of body</p>`)
	ts.NoError(err)
}

func (ts *TestSuite) TestRawEmail() {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	defer log.SetOutput(os.Stdout)

	raw := rawEmail(
		"to@example.com",
		domain.Env.EmailFromAddress,
		"test subject",
		`<h4>body</h4><img src="cid:logo"><p>End of body</p>`)

	ts.Greater(len(raw), 1000)

	ts.Equal("", buf.String(), "Got an unexpected error log entry")
}
