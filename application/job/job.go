package job

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

const (
	NewThreadMessage = "new_thread_message"
)

var w worker.Worker

func init() {
	w = worker.NewSimple()
	if err := w.Register(NewThreadMessage, newThreadMessageHandler); err != nil {
		domain.ErrLogger.Printf("error registering '%s' worker, %s", NewThreadMessage, err)
	}
}

// newThreadMessageHandler is the Worker handler for new notifications of new Thread Messages
func newThreadMessageHandler(args worker.Args) error {
	id, ok := args[domain.ArgMessageID].(int)
	if !ok || id <= 0 {
		return fmt.Errorf("no message ID provided to %s worker, args = %+v", NewThreadMessage, args)
	}

	var m models.Message
	if err := m.FindByID(id, "SentBy", "Thread"); err != nil {
		return fmt.Errorf("bad ID (%d) received by new thread message handler, %s", id, err)
	}

	if err := m.Thread.Load("Participants", "Post"); err != nil {
		return errors.New("failed to load Participants and Post in new thread message handler")
	}

	msg := notifications.Message{
		Template: domain.MessageTemplateNewThreadMessage,
		Data: map[string]interface{}{
			"appName":        domain.Env.AppName,
			"uiURL":          domain.Env.UIURL,
			"postURL":        domain.GetPostUIURL(m.Thread.Post.UUID.String()),
			"postTitle":      m.Thread.Post.Title,
			"messageContent": m.Content,
			"sentByNickname": m.SentBy.Nickname,
			"threadURL":      domain.GetThreadUIURL(m.Thread.UUID.String()),
		},
		FromEmail: domain.Env.EmailFromAddress,
	}

	var lastErr error
	for _, p := range m.Thread.Participants {
		if p.ID == m.SentBy.ID {
			continue
		}

		var tp models.ThreadParticipant
		if err := tp.FindByThreadIDAndUserID(m.ThreadID, p.ID); err != nil {
			domain.ErrLogger.Printf("newThreadMessageHandler error, %s", err)
			lastErr = err
			continue
		}
		// Don't send a notification if this user has viewed the message or if they've already been notified
		if tp.LastViewedAt.After(m.UpdatedAt) || tp.LastNotifiedAt.After(m.UpdatedAt) {
			continue
		}

		msg.ToName = p.Nickname
		msg.ToEmail = p.Email
		if err := notifications.Send(msg); err != nil {
			domain.ErrLogger.Printf("error sending 'New Thread Message' notification, %s", err)
			lastErr = err
			continue
		}

		if err := tp.UpdateLastNotifiedAt(time.Now()); err != nil {
			domain.ErrLogger.Printf("newThreadMessageHandler error, %s", err)
			lastErr = err
		}
	}

	return lastErr
}

// SubmitDelayed enqueues a new Worker job for the given handler. Arguments can be provided in `args`.
func SubmitDelayed(handler string, delay time.Duration, args map[string]interface{}) error {
	job := worker.Job{
		Queue:   "default",
		Args:    args,
		Handler: handler,
	}
	if err := w.PerformIn(job, delay); err != nil {
		domain.ErrLogger.Print(err)
		return err
	}

	return nil
}
