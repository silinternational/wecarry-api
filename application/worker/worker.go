package worker

import (
	"errors"
	"fmt"
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

func NewMessageWorker(args worker.Args) error {
	domain.Logger.Printf("--------- new_message worker, args: %+v", args)

	id, ok := args["id"]
	if !ok {
		err := errors.New("no ID provided to new_message worker")
		domain.ErrLogger.Print(err)
		return err
	}

	var m models.Message
	if err := models.DB.Find(&m, id); err != nil {
		err := fmt.Errorf("bad ID provided to new_message worker: %v", id)
		domain.ErrLogger.Print(err)
		return err
	}

	var tp models.ThreadParticipant
	if err := models.DB.Where("thread_id = ? AND user_id = ?", m.ThreadID, m.???).First(&tp); err != nil {
		return err
	}

	return nil
}
