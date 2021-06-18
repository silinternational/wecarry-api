package actions

import (
	"errors"
	"fmt"
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

	threads, err := cUser.GetThreads()
	if err != nil {
		return reportError(c, &apitypes.AppError{
			HttpStatus: http.StatusInternalServerError,
			Key:        wcerror.ThreadsLoadFailure,
			DebugMsg:   err.Error(),
		})
	}

	fmt.Printf("\nUSER THREADS: %+v\n", threads)

	var output apitypes.Threads
	if err := domain.ConvertToOtherType(threads, &output); err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	// Hydrate the thread's messages, participants and request
	for i := range threads {
		messages, err := threads[i].Messages()
		if err != nil {
			err = errors.New("error retrieving threads messages: " + err.Error())
			return reportError(c, appErrorFromErr(err))
		}

		var messagesOutput apitypes.Messages
		if err := domain.ConvertToOtherType(messages, &messagesOutput); err != nil {
			err = errors.New("error converting threads messages: " + err.Error())
			return reportError(c, appErrorFromErr(err))
		}

		for j := range messages {
			sender, err := messages[j].GetSender()
			if err != nil {
				err = errors.New("error retrieving threads message sender: " + err.Error())
				return reportError(c, appErrorFromErr(err))
			}

			var senderOutput apitypes.User
			if err := domain.ConvertToOtherType(sender, &senderOutput); err != nil {
				err = errors.New("error converting threads message sender: " + err.Error())
				return reportError(c, appErrorFromErr(err))
			}
			messagesOutput[j].Sender = senderOutput
		}

		output[i].Messages = messagesOutput

		participants, err := threads[i].GetParticipants()
		if err != nil {
			err := errors.New("error retrieving threads participants: " + err.Error())
			return reportError(c, appErrorFromErr(err))
		}

		var participantsOutput apitypes.Users
		if err := domain.ConvertToOtherType(participants, &participantsOutput); err != nil {
			err := errors.New("error converting threads participants: " + err.Error())
			return reportError(c, appErrorFromErr(err))
		}

		output[i].Participants = participantsOutput

		request, err := threads[i].GetRequest()
		if err != nil {
			err := errors.New("error retrieving threads request: " + err.Error())
			return reportError(c, appErrorFromErr(err))
		}

		var requestOutput apitypes.Request
		if err := domain.ConvertToOtherType(request, &requestOutput); err != nil {
			err := errors.New("error converting threads request: " + err.Error())
			return reportError(c, appErrorFromErr(err))
		}

		reqCreatedBy, err := request.GetCreator()
		if err != nil {
			err := errors.New("error getting threads requests created_by: " + err.Error())
			return reportError(c, appErrorFromErr(err))
		}

		var reqCreatedByOutput apitypes.User
		if err := domain.ConvertToOtherType(reqCreatedBy, &reqCreatedByOutput); err != nil {
			err := errors.New("error converting threads requests created_by: " + err.Error())
			return reportError(c, appErrorFromErr(err))
		}

		requestOutput.CreatedBy = reqCreatedByOutput

		output[i].Request = requestOutput
	}

	return c.Render(200, render.JSON(output))
}
