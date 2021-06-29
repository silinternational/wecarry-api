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

// swagger:operation POST /messages Messages CreateMessage
//
// Create a new message on a request's thread
//
// ---
// parameters:
//   - name: message
//     in: body
//     description: message input object
//     required: true
//     schema:
//       "$ref": "#/definitions/MessageInput"
// responses:
//   '200':
//     description: conversation/thread that the new message was added to
//     schema:
//       "$ref": "#/definitions/Thread"
func messagesCreate(c buffalo.Context) error {

	var input api.MessageInput
	if err := StrictBind(c, &input); err != nil {
		return reportError(c, &api.AppError{
			HttpStatus: http.StatusBadRequest,
			Key:        api.InvalidRequestBody,
			Err:        errors.New("unable to unmarshal Message data into MessageInput struct, error: " + err.Error()),
		})
	}
	user := models.CurrentUser(c)
	tx := models.Tx(c)
	var message models.Message

	if appErr := message.CreateForRest(tx, user, input); appErr != nil {
		appErr.HttpStatus = httpStatusForErrCategory(appErr.Category)
		return reportError(c, appErr)
	}

	if err := message.Thread.LoadForAPI(tx, user); err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	output, err := convertThread(c, message.Thread)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(output))
}

func convertMessagesToAPIType(ctx context.Context, messages models.Messages) (api.Messages, error) {
	var output api.Messages
	if err := api.ConvertToOtherType(messages, &output); err != nil {
		err = errors.New("error converting messages to api.Messages: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages with their sentBy users
	for i := range output {
		sentByOutput, err := convertUser(ctx, messages[i].SentBy)
		if err != nil {
			err = errors.New("error converting messages sentBy to api.User: " + err.Error())
			return nil, err
		}

		output[i].ID = messages[i].UUID
		output[i].SentBy = &sentByOutput
	}

	return output, nil
}
