package conversions

import (
	"context"
	"errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func ConvertMessagesToAPIType(ctx context.Context, messages models.Messages) (api.Messages, error) {
	var output api.Messages
	if err := api.ConvertToOtherType(messages, &output); err != nil {
		err = errors.New("error converting messages to api.Messages: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages with their sentBy users
	for i := range output {
		sentByOutput, err := ConvertUser(ctx, messages[i].SentBy)
		if err != nil {
			err = errors.New("error converting messages sentBy to api.User: " + err.Error())
			return nil, err
		}

		output[i].ID = messages[i].UUID
		output[i].SentBy = &sentByOutput
	}

	return output, nil
}
