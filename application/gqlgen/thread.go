package gqlgen

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// ThreadFields maps GraphQL fields to their equivalent database fields. For related types, the
// foreign key field name is provided.
func ThreadFields() map[string]string {
	return map[string]string{
		"id":        "uuid",
		"postID":    "post_id",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	}
}

// Thread is required by gqlgen
func (r *Resolver) Thread() ThreadResolver {
	return &threadResolver{r}
}

type threadResolver struct{ *Resolver }

// Participants resolves the `participants` property of the thread query, retrieving the related records from the
// database.
func (r *threadResolver) Participants(ctx context.Context, obj *models.Thread) ([]models.User, error) {
	if obj == nil {
		return nil, nil
	}

	participants, err := obj.GetParticipants()
	if err != nil {
		return nil, reportError(ctx, err, "GetThreadParticipants")
	}

	return participants, nil
}

// ID resolves the `ID` property of the thread query. It provides the UUID instead of the autoincrement ID.
func (r *threadResolver) ID(ctx context.Context, obj *models.Thread) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

// LastViewedAt retrieves the last_viewed_at field for the current user on the thread
func (r *threadResolver) LastViewedAt(ctx context.Context, obj *models.Thread) (*time.Time, error) {
	if obj == nil {
		return nil, nil
	}

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	lastViewedAt, err := obj.GetLastViewedAt(currentUser)
	if err != nil {
		extras := map[string]interface{}{
			"user": currentUser.Uuid,
		}
		return nil, reportError(ctx, err, "GetThreadLastViewedAt", extras)
	}

	return lastViewedAt, nil
}

// Messages resolves the `messages` property of the thread query, retrieving the related records from the
// database.
func (r *threadResolver) Messages(ctx context.Context, obj *models.Thread) ([]models.Message, error) {
	if obj == nil {
		return nil, nil
	}

	messages, err := obj.GetMessages()
	if err != nil {
		return nil, reportError(ctx, err, "GetThreadMessages")
	}

	return messages, nil
}

// PostID retrieves the UUID of the post to which the queried thread belongs.
func (r *threadResolver) PostID(ctx context.Context, obj *models.Thread) (string, error) {
	if obj == nil {
		return "", nil
	}

	post, err := obj.GetPost()
	if err != nil {
		return "", reportError(ctx, err, "GetThreadPostID")
	}

	return post.Uuid.String(), nil
}

// Post retrieves the post to which the queried thread belongs.
func (r *threadResolver) Post(ctx context.Context, obj *models.Thread) (*models.Post, error) {
	if obj == nil {
		return nil, nil
	}

	post, err := obj.GetPost()
	if err != nil {
		return nil, reportError(ctx, err, "GetThreadPost")
	}

	return post, nil
}

// UnreadMessageCount retrieves the number of unread messages the current user has on the queried thread.
func (r *threadResolver) UnreadMessageCount(ctx context.Context, obj *models.Thread) (int, error) {
	if obj == nil {
		return 0, nil
	}
	user := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	lastViewedAt, err := obj.GetLastViewedAt(user)
	if err != nil {
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
		return 0, nil
	}

	if lastViewedAt == nil {
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx),
			fmt.Sprintf("lastViewedAt nil for user %v on thread %v", user.ID, obj.ID))
		return 0, nil
	}

	count, err2 := obj.UnreadMessageCount(user.ID, *lastViewedAt)
	if err2 != nil {
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err2.Error())
		return 0, nil
	}
	return count, nil
}

// Threads retrieves the all of the threads
func (r *queryResolver) Threads(ctx context.Context) ([]models.Thread, error) {
	threads := models.Threads{}

	if err := threads.All(); err != nil {
		return nil, reportError(ctx, err, "GetThreads")
	}

	return threads, nil
}

// MyThreads retrieves all of the threads for the current user
func (r *queryResolver) MyThreads(ctx context.Context) ([]models.Thread, error) {
	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	threads, err := currentUser.GetThreads()
	if err != nil {
		extras := map[string]interface{}{
			"user": currentUser.Uuid,
		}
		return nil, reportError(ctx, err, "GetMyThreads", extras)
	}

	return threads, nil
}

func getSelectFieldsForThreads(ctx context.Context) []string {
	selectFields := GetSelectFieldsFromRequestFields(ThreadFields(), graphql.CollectAllFields(ctx))

	selectFields = append(selectFields, "id")

	return selectFields
}
