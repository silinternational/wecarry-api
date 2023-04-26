package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/job"
	"github.com/silinternational/wecarry-api/log"
)

// ServiceInput defines the input parameters to the "service" endpoint
type ServiceInput struct {
	Task ServiceTaskName `json:"task"`
}

// ServiceTask is a type of task to be issued by the "service" endpoint
type ServiceTask struct {
	Handler ServiceTaskHandler
}

// ServiceTaskName is the name of a type of task to be issued by the "service" endpoint
type ServiceTaskName string

// ServiceTaskHandler is a handler function that is executed when a specified task is requested by the API client
type ServiceTaskHandler func(buffalo.Context) error

const (
	// ServiceTaskFileCleanup removes files not linked to any object
	ServiceTaskFileCleanup ServiceTaskName = job.FileCleanup

	// ServiceTaskLocationCleanup removes locations not used by any object
	ServiceTaskLocationCleanup ServiceTaskName = job.LocationCleanup

	// ServiceTaskTokenCleanup removes expired user access tokens
	ServiceTaskTokenCleanup ServiceTaskName = job.TokenCleanup

	// ServiceTaskOutdatedRequests sends emails to users who have requests with an outdated needed_before
	ServiceTaskOutdatedRequests ServiceTaskName = job.OutdatedRequests
)

var serviceTasks = map[ServiceTaskName]ServiceTask{
	ServiceTaskFileCleanup: {
		Handler: fileCleanupHandler,
	},
	ServiceTaskLocationCleanup: {
		Handler: locationCleanupHandler,
	},
	ServiceTaskTokenCleanup: {
		Handler: tokenCleanupHandler,
	},
	ServiceTaskOutdatedRequests: {
		Handler: outdatedRequestsHandler,
	},
}

func serviceHandler(c buffalo.Context) error {
	if domain.Env.ServiceIntegrationToken == "" {
		return c.Error(http.StatusInternalServerError, errors.New("no ServiceIntegrationToken configured"))
	}

	bearerToken := domain.GetBearerTokenFromRequest(c.Request())
	if domain.Env.ServiceIntegrationToken != bearerToken {
		return c.Error(http.StatusUnauthorized, errors.New("incorrect bearer token provided"))
	}

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.Error(http.StatusInternalServerError, fmt.Errorf("error reading request body, %s", err))
	}

	var input ServiceInput
	if err := json.Unmarshal(body, &input); err != nil {
		return c.Error(http.StatusBadRequest, fmt.Errorf("error parsing request body, %s", err))
	}

	log.Errorf("scheduling service task '%s'", input.Task)

	if task, ok := serviceTasks[input.Task]; ok {
		if err := task.Handler(c); err != nil {
			return c.Error(http.StatusInternalServerError, fmt.Errorf("task %s failed, %s", input.Task, err))
		}
		return c.Render(http.StatusNoContent, nil)
	}
	return c.Error(http.StatusUnprocessableEntity, fmt.Errorf("invalid task name: %s", input.Task))
}

func fileCleanupHandler(c buffalo.Context) error {
	if err := job.Submit(job.FileCleanup, nil); err != nil {
		return c.Error(http.StatusInternalServerError, fmt.Errorf("file cleanup job not started, %s", err))
	}
	return nil
}

func locationCleanupHandler(c buffalo.Context) error {
	if err := job.Submit(job.LocationCleanup, nil); err != nil {
		return c.Error(http.StatusInternalServerError, fmt.Errorf("location cleanup job not started, %s", err))
	}
	return nil
}

func tokenCleanupHandler(c buffalo.Context) error {
	if err := job.Submit(job.TokenCleanup, nil); err != nil {
		return c.Error(http.StatusInternalServerError, fmt.Errorf("token cleanup job not started, %s", err))
	}
	return nil
}

func outdatedRequestsHandler(c buffalo.Context) error {
	if err := job.Submit(job.OutdatedRequests, nil); err != nil {
		return c.Error(http.StatusInternalServerError, fmt.Errorf("outdated requests job not started, %s", err))
	}
	return nil
}
