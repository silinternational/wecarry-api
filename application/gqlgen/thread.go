package gqlgen

import (
	"context"
	"fmt"
	"time"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// Thread is required by gqlgen
func (r *Resolver) Thread() ThreadResolver {
	return &threadResolver{r}
}

type threadResolver struct{ *Resolver }

// Participants resolves the `participants` property of the thread query, retrieving the related records from the
// database.
func (r *threadResolver) Participants(ctx context.Context, obj *models.Thread) ([]PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	participants, err := obj.GetParticipants(models.Tx(ctx))
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetThreadParticipants")
	}

	return getPublicProfiles(ctx, participants), nil
}

// ID resolves the `ID` property of the thread query. It provides the UUID instead of the autoincrement ID.
func (r *threadResolver) ID(ctx context.Context, obj *models.Thread) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// LastViewedAt retrieves the last_viewed_at field for the current user on the thread
func (r *threadResolver) LastViewedAt(ctx context.Context, obj *models.Thread) (*time.Time, error) {
	if obj == nil {
		return nil, nil
	}

	currentUser := models.CurrentUser(ctx)
	lastViewedAt, err := obj.GetLastViewedAt(ctx, currentUser)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetThreadLastViewedAt")
	}

	return lastViewedAt, nil
}

// Messages resolves the `messages` property of the thread query, retrieving the related records from the
// database.
func (r *threadResolver) Messages(ctx context.Context, obj *models.Thread) ([]models.Message, error) {
	if obj == nil {
		return nil, nil
	}

	messages, err := obj.Messages(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetThreadMessages")
	}

	return messages, nil
}

// Request retrieves the request to which the queried thread belongs.
func (r *threadResolver) Request(ctx context.Context, obj *models.Thread) (*models.Request, error) {
	if obj == nil {
		return &models.Request{}, nil
	}

	request, err := obj.GetRequest(models.Tx(ctx))
	if err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "GetThreadRequest")
	}

	return request, nil
}

// UnreadMessageCount retrieves the number of unread messages the current user has on the queried thread.
func (r *threadResolver) UnreadMessageCount(ctx context.Context, obj *models.Thread) (int, error) {
	if obj == nil {
		return 0, nil
	}
	user := models.CurrentUser(ctx)

	lastViewedAt, err := obj.GetLastViewedAt(ctx, user)
	if err != nil {
		domain.Warn(ctx, err.Error())
		return 0, nil
	}

	if lastViewedAt == nil {
		domain.Warn(ctx,
			fmt.Sprintf("lastViewedAt nil for user %v on thread %v", user.ID, obj.ID))
		return 0, nil
	}

	count, err2 := obj.UnreadMessageCount(ctx, user.ID, *lastViewedAt)
	if err2 != nil {
		domain.Warn(ctx, err2.Error())
		return 0, nil
	}
	return count, nil
}

// Threads retrieves all of the threads for the current user. It is a placeholder for an admin tool for retrieving
// all threads for a specified user.
func (r *queryResolver) Threads(ctx context.Context) ([]models.Thread, error) {
	currentUser := models.CurrentUser(ctx)

	threads, err := currentUser.GetThreads(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetThreads")
	}

	return threads, nil
}

// MyThreads retrieves all of the threads for the current user
func (r *queryResolver) MyThreads(ctx context.Context) ([]models.Thread, error) {
	currentUser := models.CurrentUser(ctx)

	threads, err := currentUser.GetThreads(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetMyThreads")
	}

	return threads, nil
}
