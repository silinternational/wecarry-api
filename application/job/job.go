package job

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/silinternational/wecarry-api/notifications"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

var W worker.Worker

func init() {
	W = worker.NewSimple()
	if err := W.Register("new_message", NewMessageWorker); err != nil {
		domain.ErrLogger.Printf("error registering 'new_message' worker, %s", err)
	}
}

// NewMessageWorker is the Worker handler for new notifications of new Thread Messages
func NewMessageWorker(args worker.Args) error {
	domain.Logger.Printf("--------- new_message worker, args: %+v", args)

	id, ok := args["message_id"].(int)
	if !ok {
		err := errors.New("no message ID provided to new_message worker")
		domain.ErrLogger.Print(err)
		return err
	}

	var m models.Message
	if err := m.FindByID(id, "SentBy", "Thread"); err != nil {
		return fmt.Errorf("sendNewMessageNotification: bad ID (%d) received in event payload, %s", id, err)
	}

	if err := m.Thread.LoadRelations("Participants", "Post"); err != nil {
		return fmt.Errorf("sendNewMessageNotification: failed to load Participants and Post")
	}

	var recipients []struct{ Nickname, Email string }
	for _, p := range m.Thread.Participants {
		if p.ID == m.SentBy.ID {
			continue
		}

		var tp models.ThreadParticipant
		if err := models.DB.Where("user_id = ? AND thread_id = ?", p.ID, m.ThreadID).First(&tp); err != nil {
			return fmt.Errorf("failed to find thread_participant record for user %d and thread %d", tp.ID, m.ThreadID)
		}
		if tp.LastViewedAt.After(tp.LastNotifiedAt) {
			continue
		}

		tp.LastNotifiedAt = time.Now()
		if err := models.DB.Update(&tp); err != nil {
			return errors.New("failed to update thread_participant.last_notified_at")
		}

		recipients = append(recipients,
			struct{ Nickname, Email string }{p.Nickname, p.Email})
	}

	uiUrl := envy.Get(domain.UIURLEnv, "")
	data := map[string]interface{}{
		"postURL":        uiUrl + "/#/requests/" + m.Thread.Post.Uuid.String(),
		"postTitle":      m.Thread.Post.Title,
		"messageContent": m.Content,
		"sentByNickname": m.SentBy.Nickname,
		"threadURL":      uiUrl + "/#/messages/" + m.Thread.Uuid.String(),
	}

	for _, r := range recipients {
		msg := notifications.Message{
			Template:  domain.MessageTemplateNewMessage,
			Data:      data,
			FromName:  m.SentBy.Nickname,
			FromEmail: m.SentBy.Email,
			ToName:    r.Nickname,
			ToEmail:   r.Email,
		}
		if err := notifications.Send(msg); err != nil {
			domain.ErrLogger.Printf("error sending 'New Message' notification, %s", err)
		}
	}

	return nil
}

// Submit enqueues a new Worker job for the given handler. Arguments can be provided in `args`.
func Submit(handler string, args map[string]interface{}) error {
	job := worker.Job{
		Queue:   "default",
		Args:    args,
		Handler: handler,
	}
	if err := W.PerformIn(job, 10*time.Second); err != nil {
		domain.ErrLogger.Print(err)
		return err
	}

	return nil
}
