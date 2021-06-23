package actions

import (
	"context"
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

// usersThreads responds to Get requests at /conversations
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

	output, err := convertThreadsToAPIType(c, threads)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(output))
}

// swagger:operation POST /projects Threads SetThreadLastViewedAt
//
// Sets the last viewed time for the current user on the given thread
//
// ---
//   - name: set_thread_last_viewed_at
//     in: body
//     description: SetThreadLastViewedAt input object
//     schema:
//       "$ref": "#/definitions/SetThreadLastViewedAtInput"
func threadsMarkMessagesAsRead(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	var thread models.Thread

	var input api.SetThreadLastViewedAtInput
	if err := StrictBind(c, &input); err != nil {
		return reportError(c, &api.AppError{
			HttpStatus: http.StatusBadRequest,
			Key:        api.InvalidRequestBody,
			Err:        errors.New("unable to unmarshal Post data into SetThreadLastViewedAtInput struct, error: " + err.Error()),
		})
	}

	if err := thread.FindByUUID(tx, input.ThreadID); err != nil {
		return reportError(c, &api.AppError{
			HttpStatus: http.StatusNotFound,
			Key:        api.ThreadNotFound,
			Err:        err,
		})
	}

	if err := thread.UpdateLastViewedAt(tx, cUser.ID, input.Time); err != nil {
		return reportError(c, &api.AppError{
			HttpStatus: http.StatusInternalServerError,
			Key:        api.ThreadSetLastViewedAt,
			Err:        err,
		})
	}

	if err := thread.LoadForAPI(tx, cUser); err != nil {
		return reportError(c, &api.AppError{
			HttpStatus: http.StatusInternalServerError,
			Key:        api.ThreadsLoadFailure,
			Err:        err,
		})
	}

	threads := models.Threads{thread}
	converted, err := convertThreadsToAPIType(c, threads)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	// this should never happen, but just in case ...
	if len(converted) == 0 {
		return reportError(c, appErrorFromErr(errors.New("thread got lost in conversion")))
	}

	return c.Render(200, render.JSON(converted[0]))
}

func convertThreadsToAPIType(c context.Context, threads models.Threads) (api.Threads, error) {
	var output api.Threads
	if err := api.ConvertToOtherType(threads, &output); err != nil {
		err = errors.New("error converting threads to api.threads: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages, participants
	for i := range threads {
		messagesOutput, err := convertMessagesToAPIType(c, threads[i].Messages)
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
		output[i].ID = threads[i].UUID
	}

	return output, nil
}
