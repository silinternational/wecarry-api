package actions

import (
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

// uploadHandler responds to POST requests at /upload
func usersThreads(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	threads, err := cUser.GetThreadsForConversations(tx)
	if err != nil {
		return reportError(c, &api.AppError{
			HttpStatus: http.StatusInternalServerError,
			Key:        api.ThreadsLoadFailure,
			Err:        err,
		})
	}

	output, err := convertThreadsToAPIType(threads)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(output))
}

func convertThreadsToAPIType(threads models.Threads) (api.Threads, error) {
	var output api.Threads
	if err := api.ConvertToOtherType(threads, &output); err != nil {
		err = errors.New("error converting threads to apitype.threads: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages, participants
	for i := range threads {
		messagesOutput, err := convertMessagesToAPIType(threads[i].Messages)
		if err != nil {
			return nil, err
		}
		output[i].Messages = &messagesOutput

		// Not converting Participants, since that happens automatically  above and
		// because it doesn't have nested related objects

		requestOutput, err := convertRequestToAPIType(threads[i].Request)
		if err != nil {
			return nil, err
		}

		output[i].Request = &requestOutput
	}

	return output, nil
}
