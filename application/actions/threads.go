package actions

import (
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/apitypes"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/wcerror"
)

// uploadHandler responds to POST requests at /upload
func usersThreads(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	threads, err := cUser.GetThreadsForConversations(tx)
	if err != nil {
		return reportError(c, &apitypes.AppError{
			HttpStatus: http.StatusInternalServerError,
			Key:        wcerror.ThreadsLoadFailure,
			DebugMsg:   err.Error(),
		})
	}

	output, err := convertThreadsToAPIType(threads)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(output))
}

func convertThreadsToAPIType(threads models.Threads) (apitypes.Threads, error) {
	var output apitypes.Threads
	if err := domain.ConvertToOtherType(threads, &output); err != nil {
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
