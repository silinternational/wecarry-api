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
	FileCleanup      = "file_cleanup"
	TokenCleanup     = "token_cleanup"
)

var w worker.Worker

func init() {
	w = worker.NewSimple()
	if err := w.Register(NewThreadMessage, newThreadMessageHandler); err != nil {
		domain.ErrLogger.Printf("error registering '%s' worker, %s", NewThreadMessage, err)
	}
	if err := w.Register(FileCleanup, fileCleanupHandler); err != nil {
		domain.ErrLogger.Printf("error registering '%s' worker, %s", FileCleanup, err)
	}
	if err := w.Register(TokenCleanup, tokenCleanupHandler); err != nil {
		domain.ErrLogger.Printf("error registering '%s' worker, %s", TokenCleanup, err)
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

	template := domain.MessageTemplateNewThreadMessage
	msg := notifications.Message{
		Template: template,
		Data: map[string]interface{}{
			"appName":        domain.Env.AppName,
			"uiURL":          domain.Env.UIURL,
			"postURL":        domain.GetPostUIURL(m.Thread.Post.UUID.String()),
			"postTitle":      domain.Truncate(m.Thread.Post.Title, "...", 16),
			"messageContent": m.Content,
			"sentByNickname": m.SentBy.Nickname,
			"threadURL":      domain.GetThreadUIURL(m.Thread.UUID.String()),
		},
		FromEmail: domain.EmailFromAddress(&m.SentBy.Nickname),
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

		msg.ToName = p.GetRealName()
		msg.ToEmail = p.Email
		msg.Subject = domain.GetTranslatedSubject(p.GetLanguagePreference(),
			"Email.Subject.Message.Created",
			map[string]string{"sentByNickname": m.SentBy.Nickname})

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

// fileCleanupHandler removes unlinked files
func fileCleanupHandler(args worker.Args) error {
	files := models.Files{}
	if err := files.DeleteUnlinked(); err != nil {
		return fmt.Errorf("file cleanup failed with error, %s", err)
	}
	return nil
}

// tokenCleanupHandler removes expired user access tokens
func tokenCleanupHandler(args worker.Args) error {
	u := models.UserAccessTokens{}
	deleted, err := u.DeleteExpired()
	if err != nil {
		return fmt.Errorf("error cleaning expired user access tokens: %v", err)
	}

	domain.Logger.Printf("Deleted %v expired user access tokens during cleanup", deleted)
	return nil
}

// SubmitDelayed enqueues a new Worker job for the given handler. Arguments can be provided in `args`.
func SubmitDelayed(handler string, delay time.Duration, args map[string]interface{}) error {
	job := worker.Job{
		Queue:   "default",
		Args:    args,
		Handler: handler,
	}
	return w.PerformIn(job, delay)
}
