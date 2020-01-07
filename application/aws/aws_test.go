package aws

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/silinternational/wecarry-api/domain"
)

// TestSuite establishes a test suite
type TestSuite struct {
	suite.Suite
}

// Test_TestSuite runs the test suite
func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// TestSendEmail can be used in local development if the environment is configured with SES credentials
//func (ts *TestSuite) TestSendEmail() {
//	err := SendEmail(
//		"me@example.com",
//		domain.Env.EmailFromAddress,
//		"test subject",
//		`<h4>body</h4><img src="cid:logo"><p>End of body</p>`)
//	ts.NoError(err)
//}

func (ts *TestSuite) TestRawEmail() {
	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer domain.ErrLogger.SetOutput(os.Stderr)

	raw := rawEmail(
		"to@example.com",
		domain.Env.EmailFromAddress,
		"test subject",
		`<h4>body</h4><img src="cid:logo"><p>End of body</p>`)

	ts.Greater(len(raw), 1000)

	ts.Equal("", buf.String(), "Got an unexpected error log entry")
}
