package gqlgen

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
	"strconv"
)

func ThreadSimpleFields() map[string]string {
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

	selectedFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), GetRequestFields(ctx))
	participants, err := models.GetThreadParticipants(obj.ID, selectedFields)
	if err != nil {
		return nil, err
	}

	var users []*models.User

	for _, p := range participants {
		// TODO: change models.GetThreadParticipants to return a slice of pointers
		u := p // instantiate a new copy for each object
		users = append(users, &u)
	}

	return users, nil
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
	return nil, nil
}

func (r *threadResolver) PostID(ctx context.Context, obj *models.Thread) (string, error) {
	if obj == nil {
		return "", nil
	}
	return strconv.Itoa(obj.PostID), nil
}

func (r *threadResolver) CreatedAt(ctx context.Context, obj *models.Thread) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *threadResolver) UpdatedAt(ctx context.Context, obj *models.Thread) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}

func (r *queryResolver) Threads(ctx context.Context) ([]*models.Thread, error) {
	var threads []*models.Thread

	db := models.DB

	requestFields := GetRequestFields(ctx)
	selectFields := getSelectFieldsForThreads(requestFields)

	if err := db.Select(selectFields...).All(&threads); err != nil {
		return []*models.Thread{}, fmt.Errorf("error getting threads: %v", err)
	}

	// TODO: extract common code between Threads and MyThreads
	// TODO: filter requestFields for post fields
	postFields := requestFields
	messageFields := GetSelectFieldsFromRequestFields(MessageSimpleFields(), requestFields)
	for _, t := range threads {
		if err := t.LoadPost(postFields); err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error loading post data for thread %v: %v", t.ID, err))
		}

		if err := t.LoadMessages(messageFields); err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error loading messages for thread %v: %v", t.ID, err))
		}
	}

	return threads, nil
}

func (r *queryResolver) MyThreads(ctx context.Context) ([]*models.Thread, error) {
	var threads []*models.Thread

	db := models.DB
	requestFields := GetRequestFields(ctx)
	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	// TODO: use getSelectFieldsForThreads()
	query := db.Q().LeftJoin("thread_participants tp", "threads.id = tp.thread_id")
	query = query.Where("tp.user_id = ?", currentUser.ID)
	if err := query.All(&threads); err != nil {
		return []*models.Thread{}, fmt.Errorf("error getting threads: %v", err)
	}

	// TODO: extract common code between Threads and MyThreads
	// TODO: filter requestFields for post fields
	postFields := requestFields
	messageFields := GetSelectFieldsFromRequestFields(MessageSimpleFields(), requestFields)
	for _, t := range threads {
		if err := t.LoadPost(postFields); err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error loading post data for thread %v: %v", t.ID, err))
		}

		if err := t.LoadMessages(messageFields); err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error loading messages for thread %v: %v", t.ID, err))
		}
	}

	return threads, nil
}

func getSelectFieldsForThreads(requestFields []string) []string {
	selectFields := GetSelectFieldsFromRequestFields(ThreadSimpleFields(), requestFields)

	// Ensure we can get participants via the thread ID
	if domain.IsStringInSlice(ParticipantsField, requestFields) {
		selectFields = append(selectFields, "id")
	}

	// Ensure we can get the post via the post ID
	if domain.IsStringInSlice(PostField, requestFields) {
		selectFields = append(selectFields, "post_id")
	}

	return selectFields
}
