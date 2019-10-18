package job

import (
	"errors"
	"fmt"
	"time"

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

	id, ok := args["message_id"]
	if !ok {
		err := errors.New("no message ID provided to new_message worker")
		domain.ErrLogger.Print(err)
		return err
	}

	var m models.Message
	if err := models.DB.Find(&m, id); err != nil {
		err := fmt.Errorf("bad ID provided to new_message worker: %v", id)
		domain.ErrLogger.Print(err)
		return err
	}

	return nil
}

func Submit(handler string, args map[string]interface{}) error {
	job := worker.Job{
		Queue:   "default",
		Args:    args,
		Handler: handler,
	}
	if err := W.PerformIn(job, 10*time.Second); err != nil {
		domain.ErrLogger.Print(err)
	}

	return nil
}
