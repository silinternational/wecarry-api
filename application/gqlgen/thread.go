package gqlgen

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func ThreadFields() map[string]string {
	return map[string]string{
		"id":        "uuid",
		"postID":    "post_id",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	}
}

func (r *Resolver) Thread() ThreadResolver {
	return &threadResolver{r}
}

type threadResolver struct{ *Resolver }

func (r *threadResolver) Participants(ctx context.Context, obj *models.Thread) ([]*models.User, error) {
	if obj == nil {
		return nil, nil
	}

	selectedFields := GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx))
	return obj.GetParticipants(selectedFields)
}

func (r *threadResolver) ID(ctx context.Context, obj *models.Thread) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

func (r *threadResolver) Messages(ctx context.Context, obj *models.Thread) ([]*models.Message, error) {
	if obj == nil {
		return nil, nil
	}
	selectedFields := GetSelectFieldsFromRequestFields(MessageFields(), graphql.CollectAllFields(ctx))
	return obj.GetMessages(selectedFields)
}

func (r *threadResolver) PostID(ctx context.Context, obj *models.Thread) (string, error) {
	if obj == nil {
		return "", nil
	}
	if post, err := obj.GetPost([]string{"uuid"}); err == nil {
		return post.Uuid.String(), nil
	} else {
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return "", err
	}
}

func (r *threadResolver) Post(ctx context.Context, obj *models.Thread) (*models.Post, error) {
	selectedFields := GetSelectFieldsFromRequestFields(PostFields(), graphql.CollectAllFields(ctx))
	return obj.GetPost(selectedFields)
}

func (r *queryResolver) Threads(ctx context.Context) ([]*models.Thread, error) {
	var threads []*models.Thread

	db := models.DB

	selectFields := getSelectFieldsForThreads(graphql.CollectAllFields(ctx))
	if err := db.Select(selectFields...).All(&threads); err != nil {
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return []*models.Thread{}, fmt.Errorf("error getting threads: %v", err)
	}

	return threads, nil
}

func (r *queryResolver) MyThreads(ctx context.Context) ([]*models.Thread, error) {
	var threads []*models.Thread

	db := models.DB
	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	query := db.Q().LeftJoin("thread_participants tp", "threads.id = tp.thread_id")
	query = query.Where("tp.user_id = ?", currentUser.ID)
	if err := query.All(&threads); err != nil {
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return []*models.Thread{}, fmt.Errorf("error getting threads: %v", err)
	}

	return threads, nil
}

func getSelectFieldsForThreads(requestFields []string) []string {
	selectFields := GetSelectFieldsFromRequestFields(ThreadFields(), requestFields)

	selectFields = append(selectFields, "id")

	return selectFields
}
