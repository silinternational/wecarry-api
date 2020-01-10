package notifications

import (
	"testing"
	"text/template"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	nickname := "nickname"
	msg := Message{
		FromName:  "from name",
		FromEmail: domain.EmailFromAddress(&nickname),
		ToName:    "to name",
		ToEmail:   "to@example.com",
		Template:  domain.MessageTemplateNewThreadMessage,
		Data: map[string]interface{}{
			"uiURL":          "example.com",
			"appName":        "Our App",
			"postURL":        "mypost.example.com",
			"postTitle":      "My Post<script>doBadThings()</script>",
			"messageContent": "I can bring it<script>doBadThings()</script>",
			"sentByNickname": "Fred<script>doBadThings()</script>",
			"threadURL":      "ourthread.example.com",
		},
	}
	var emailService EmailService
	var testService DummyEmailService
	emailService = &testService

	if err := emailService.Send(msg); err != nil {
		t.Errorf("error sending message, %s", err)
	}

	n := len(testService.sentMessages)
	if n != 1 {
		t.Errorf("incorrect number of messages sent (%d)", n)
		t.FailNow()
	}

	body := testService.GetLastBody()
	assert.Contains(t, body, template.HTMLEscapeString(msg.Data["messageContent"].(string)))
	assert.NotContains(t, body, "<script>")
}
