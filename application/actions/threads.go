package actions

import (
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

	output, err := models.ConvertThreadsToAPIType(c, threads)
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
	if err = StrictBind(c, &input); err != nil {
		return reportError(c, err)
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
	converted, err := models.ConvertThreadsToAPIType(c, threads)
	if err != nil {
		return reportError(c, err)
	}

	// this should never happen, but just in case ...
	if len(converted) == 0 {
		return reportError(c, errors.New("thread got lost in conversion"))
	}

	return c.Render(200, render.JSON(converted[0]))
}
