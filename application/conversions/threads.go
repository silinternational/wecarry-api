package conversions

import (
	"context"
	"errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func ConvertThreadsToAPIType(ctx context.Context, threads models.Threads) (api.Threads, error) {
	var output api.Threads
	if err := api.ConvertToOtherType(threads, &output); err != nil {
		err = errors.New("error converting threads to api.threads: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages, participants
	for i := range output {
		messagesOutput, err := ConvertMessagesToAPIType(ctx, threads[i].Messages)
		if err != nil {
			return nil, err
		}
		output[i].Messages = &messagesOutput

		// Not converting Participants, since that happens automatically  above and
		// because it doesn't have nested related objects
		for j := range output[i].Participants {
			output[i].Participants[j].ID = threads[i].Participants[j].UUID
		}

		requestOutput, err := ConvertRequestToAPIType(ctx, threads[i].Request)
		if err != nil {
			return nil, err
		}

		output[i].Request = &requestOutput
		output[i].ID = threads[i].UUID
	}

	return output, nil
}
