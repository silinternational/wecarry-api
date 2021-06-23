package actions

import (
	"errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func convertMessagesToAPIType(messages models.Messages) (api.Messages, error) {
	var output api.Messages
	if err := api.ConvertToOtherType(messages, &output); err != nil {
		err = errors.New("error converting messages to api.messages: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages
	for i := range output {
		var sentByOutput api.User
		if err := api.ConvertToOtherType(messages[i].SentBy, &sentByOutput); err != nil {
			err = errors.New("error converting messages sent_by user: " + err.Error())
			return nil, err
		}
		output[i].SentBy = &sentByOutput
	}

	return output, nil
}
