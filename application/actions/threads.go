package actions

import (
	"context"
	"errors"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

// swagger:operation GET /threads Threads UsersThreads
//
// List the User's Conversations/Threads
//
//
// ---
// responses:
//   '200':
//     description: A list of the user's threads/conversations with their messages
//     schema:
//       "$ref": "#/definitions/Threads"
func threadsMine(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	threads, err := cUser.GetThreadsForConversations(tx)
	if err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorThreadsLoadFailure, api.CategoryInternal))
	}

	output, err := convertThreadsToAPIType(c, threads)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// swagger:operation PUT /threads/{thread_id}/read Threads MarkAsRead
//
// Sets the last viewed time for the current user on the given thread
//
// ---
// parameters:
//   - name: MarkMessagesAsReadInput
//     in: body
//     required: true
//     description: input object
//     schema:
//       "$ref": "#/definitions/MarkMessagesAsReadInput"
//
// responses:
//   '200':
//     description: A thread of messages
//     schema:
//       "$ref": "#/definitions/Thread"
func threadsMarkAsRead(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	id, err := getUUIDFromParam(c, "thread_id")
	if err != nil {
		return reportError(c, err)
	}

	var thread models.Thread

	var input api.MarkMessagesAsReadInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal Post data into MarkMessagesAsReadInput struct, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	if err := thread.FindByUUID(tx, id.String()); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorThreadNotFound, api.CategoryNotFound))
	}

	if err := thread.UpdateLastViewedAt(tx, cUser.ID, input.Time); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorThreadSetLastViewedAt, api.CategoryInternal))
	}

	if err := thread.LoadForAPI(tx, cUser); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorThreadsLoadFailure, api.CategoryInternal))
	}

	threads := models.Threads{thread}
	converted, err := convertThreadsToAPIType(c, threads)
	if err != nil {
		return reportError(c, err)
	}

	// this should never happen, but just in case ...
	if len(converted) == 0 {
		return reportError(c, errors.New("thread got lost in conversion"))
	}

	return c.Render(200, render.JSON(converted[0]))
}

func convertThreadsToAPIType(ctx context.Context, threads models.Threads) (api.Threads, error) {
	var output api.Threads
	if err := api.ConvertToOtherType(threads, &output); err != nil {
		err = errors.New("error converting threads to api.threads: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages, participants
	for i := range output {
		newThread, err := convertThread(ctx, threads[i])
		if err != nil {
			return nil, err
		}
		output[i] = newThread
	}

	return output, nil
}

func convertThread(ctx context.Context, thread models.Thread) (api.Thread, error) {
	var output api.Thread
	if err := api.ConvertToOtherType(thread, &output); err != nil {
		err = errors.New("error converting thread to api.thread: " + err.Error())
		return api.Thread{}, err
	}

	// Hydrate the thread's messages, participants

	messagesOutput, err := convertMessagesToAPIType(ctx, thread.Messages)
	if err != nil {
		return api.Thread{}, err
	}
	output.Messages = &messagesOutput

	// Not converting Participants, since that happens automatically  above and
	// because it doesn't have nested related objects
	for i := range output.Participants {
		output.Participants[i].ID = thread.Participants[i].UUID
	}

	if thread.Request.ID > 0 {
		requestOutput, err := convertRequest(ctx, thread.Request)
		if err != nil {
			return api.Thread{}, err
		}

		output.Request = &requestOutput
	} else {
		output.Request = nil
	}

	output.ID = thread.UUID
	return output, nil
}
