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
	OutdatedRequest  = "outdated_request"
	FileCleanup      = "file_cleanup"
	LocationCleanup  = "location_cleanup"
	TokenCleanup     = "token_cleanup"
)

var w *worker.Worker

var handlers = map[string]func(worker.Args) error{
	NewThreadMessage: newThreadMessageHandler,
	OutdatedRequest:  outdatedRequestMessageHandler,
	FileCleanup:      fileCleanupHandler,
	LocationCleanup:  locationCleanupHandler,
	TokenCleanup:     tokenCleanupHandler,
}

func Init(appWorker *worker.Worker) {
	w = appWorker
	for key, handler := range handlers {
		if err := (*w).Register(key, handler); err != nil {
			domain.ErrLogger.Printf("error registering '%s' handler, %s", key, err)
		}
	}
}

// outdatedRequestMessageHandler is the Worker handler for new notifications
// regarding open requests that have a needby date in the past
func outdatedRequestMessageHandler(args worker.Args) error {
	var requests models.Requests
	db := models.DB
	if err := requests.FindOpenPastNeededBefore(db, "NeededBefore", "UUID"); err != nil {
		return fmt.Errorf("error finding outdated requests for %s worker: %s", OutdatedRequest, err)
	}

	var lastErr error
	template := domain.MessageTemplateRequestPastNeededBefore
	for i, r := range requests {
		if err := db.Load(&requests[i], "CreatedBy"); err != nil {
			return fmt.Errorf("error loading CreatedBy User of request: %s", err)
		}

		requestTitle := domain.Truncate(r.Title, "...", 16)
		msg := notifications.Message{
			Template: template,
			Data: map[string]interface{}{
				"appName":        domain.Env.AppName,
				"uiURL":          domain.Env.UIURL,
				"requestURL":     domain.GetRequestUIURL(r.UUID.String()),
				"requestEditURL": domain.GetRequestEditUIURL(r.UUID.String()),
				"requestTitle":   requestTitle,
			},
			FromEmail: domain.EmailFromAddress(nil),
		}

		creator := requests[i].CreatedBy
		msg.ToName = creator.GetRealName()
		msg.ToEmail = creator.Email
		msg.Subject = domain.GetTranslatedSubject(r.CreatedBy.GetLanguagePreference(db),
			"Email.Subject.Request.Outdated",
			map[string]string{"requestTitle": requestTitle})

		if err := notifications.Send(msg); err != nil {
			domain.ErrLogger.Printf("error sending 'Outdated Request' notification, %s", err)
			lastErr = err
			continue
		}
	}

	return lastErr
}

// newThreadMessageHandler is the Worker handler for new notifications of new Thread Messages
func newThreadMessageHandler(args worker.Args) error {
	id, ok := args[domain.ArgMessageID].(int)
	if !ok || id <= 0 {
		return fmt.Errorf("no message ID provided to %s worker, args = %+v", NewThreadMessage, args)
	}

	var m models.Message
	if err := m.FindByID(models.DB, id, "SentBy", "Thread"); err != nil {
		return fmt.Errorf("bad ID (%d) received by new thread message handler, %s", id, err)
	}

	if err := m.Thread.Load(models.DB, "Participants", "Request"); err != nil {
		return errors.New("failed to load Participants and Request in new thread message handler")
	}

	template := domain.MessageTemplateNewThreadMessage
	requestTitle := domain.Truncate(m.Thread.Request.Title, "...", 16)
	msg := notifications.Message{
		Template: template,
		Data: map[string]interface{}{
			"appName":        domain.Env.AppName,
			"uiURL":          domain.Env.UIURL,
			"requestURL":     domain.GetRequestUIURL(m.Thread.Request.UUID.String()),
			"requestTitle":   requestTitle,
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
		if err := tp.FindByThreadIDAndUserID(models.DB, m.ThreadID, p.ID); err != nil {
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
		msg.Subject = domain.GetTranslatedSubject(p.GetLanguagePreference(models.DB),
			"Email.Subject.Message.Created",
			map[string]string{"sentByNickname": m.SentBy.Nickname, "requestTitle": requestTitle})

		if err := notifications.Send(msg); err != nil {
			domain.ErrLogger.Printf("error sending 'New Thread Message' notification, %s", err)
			lastErr = err
			continue
		}

		if err := tp.UpdateLastNotifiedAt(models.DB, time.Now()); err != nil {
			domain.ErrLogger.Printf("newThreadMessageHandler error, %s", err)
			lastErr = err
		}
	}

	return lastErr
}

// fileCleanupHandler removes unlinked files
func fileCleanupHandler(args worker.Args) error {
	files := models.Files{}
	if err := files.DeleteUnlinked(models.DB); err != nil {
		return fmt.Errorf("file cleanup failed with error, %s", err)
	}
	return nil
}

// locationCleanupHandler removes unused locations
func locationCleanupHandler(args worker.Args) error {
	locations := models.Locations{}
	if err := locations.DeleteUnused(); err != nil {
		return fmt.Errorf("location cleanup failed with error, %s", err)
	}
	return nil
}

// tokenCleanupHandler removes expired user access tokens
func tokenCleanupHandler(args worker.Args) error {
	u := models.UserAccessTokens{}
	deleted, err := u.DeleteExpired(models.DB)
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
	return (*w).PerformIn(job, delay)
}

// Submit enqueues a new Worker job for the given handler. Arguments can be provided in `args`.
func Submit(handler string, args map[string]interface{}) error {
	job := worker.Job{
		Queue:   "default",
		Args:    args,
		Handler: handler,
	}
	return (*w).Perform(job)
}
