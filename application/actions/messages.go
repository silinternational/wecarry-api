package actions

import (
	"errors"

	"github.com/silinternational/wecarry-api/apitypes"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func convertMessagesToAPIType(messages models.Messages) (apitypes.Messages, error) {
	var output apitypes.Messages
	if err := domain.ConvertToOtherType(messages, &output); err != nil {
		err = errors.New("error converting messages to apitype.messages: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages
	for i := range messages {
		var sentByOutput apitypes.User
		if err := domain.ConvertToOtherType(messages[i].SentBy, &sentByOutput); err != nil {
			err = errors.New("error converting messages sent_by user: " + err.Error())
			return nil, err
		}
		output[i].SentBy = &sentByOutput
	}

	return output, nil
}
