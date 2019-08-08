//go:generate go run github.com/99designs/gqlgen

package gqlgen

import (
	"context"
	"fmt"
	"strconv"

	"github.com/silinternational/handcarry-api/domain"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// TestUser is intended as a way to inject a "current User" for unit tests
var TestUser models.User

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *Resolver) Organization() OrganizationResolver {
	return &organizationResolver{r}
}

type organizationResolver struct{ *Resolver }

func (r *organizationResolver) URL(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Url), nil
}

func (r *organizationResolver) CreatedAt(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *organizationResolver) UpdatedAt(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}

func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

func (r *userResolver) CreatedAt(ctx context.Context, obj *models.User) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *userResolver) UpdatedAt(ctx context.Context, obj *models.User) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}

func (r *userResolver) AdminRole(ctx context.Context, obj *models.User) (*Role, error) {
	if obj == nil {
		return nil, nil
	}
	a := Role(obj.AdminRole.String)
	return &a, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]*models.User, error) {
	db := models.DB
	var dbUsers []*models.User

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin {
		return []*models.User{}, fmt.Errorf("not authorized")
	}

	requestFields := GetRequestFields(ctx)
	selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), requestFields)

	if err := db.Select(selectFields...).All(&dbUsers); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting users: %v", err.Error()))
		return []*models.User{}, err
	}

	return dbUsers, nil
}

func (r *queryResolver) User(ctx context.Context, id *string) (*models.User, error) {
	dbUser := models.User{}

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin && currentUser.Uuid.String() != *id {
		return &dbUser, fmt.Errorf("not authorized")
	}

	requestFields := GetRequestFields(ctx)
	selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), requestFields)

	if err := models.DB.Select(selectFields...).Where("uuid = ?", id).First(&dbUser); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting user: %v", err.Error()))
		return &dbUser, err
	}

	return &dbUser, nil
}

func (r *Resolver) Post() PostResolver {
	return &postResolver{r}
}

type postResolver struct{ *Resolver }

func (r *postResolver) Type(ctx context.Context, obj *models.Post) (PostType, error) {
	if obj == nil {
		return "", nil
	}
	return PostType(obj.Type), nil
}

func (r *postResolver) Organization(ctx context.Context, obj *models.Post) (*models.Organization, error) {
	if obj == nil {
		return nil, nil
	}
	organization, err := obj.GetOrganization(GetRequestFields(ctx))
	if err != nil {
		return nil, fmt.Errorf("error retrieving Organization data for post %v: %v", obj.ID, err)
	}
	return &organization, nil
}

func (r *postResolver) Description(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Description), nil
}

func (r *postResolver) Destination(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Destination), nil
}

func (r *postResolver) Origin(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Origin), nil
}

func (r *postResolver) NeededAfter(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.NeededAfter), nil
}

func (r *postResolver) NeededBefore(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.NeededBefore), nil
}

func (r *postResolver) Threads(ctx context.Context, obj *models.Post) ([]*models.Thread, error) {
	return nil, nil
}

func (r *postResolver) CreatedAt(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *postResolver) UpdatedAt(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}

func (r *postResolver) MyThreadID(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	thread, err := models.FindThreadByPostIDAndUserID(obj.ID, currentUser.ID)
	if err != nil {
		return nil, err
	}

	threadUuid := thread.Uuid.String()
	if threadUuid == domain.EmptyUUID {
		return nil, nil
	}

	return &threadUuid, nil
}

func (r *queryResolver) Posts(ctx context.Context) ([]*models.Post, error) {

	db := models.DB
	var posts []*models.Post
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if err := db.Where("organization_id IN (?)", cUser.GetOrgIDs()...).All(&posts); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting posts: %v", err.Error()))
		return []*models.Post{}, err
	}

	for _, p := range posts {
		e := p.QueryRelatedUsers(GetSelectFieldsFromRequestFields(UserSimpleFields(), GetRequestFields(ctx)))
		if e != nil {
			return posts, e
		}
	}

	return posts, nil
}

func (r *queryResolver) Post(ctx context.Context, id *string) (*models.Post, error) {
	post := models.Post{}
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if err := models.DB.Where("organization_id IN (?)", cUser.GetOrgIDs()...).Where("uuid = ?", id).First(&post); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting post: %v", err.Error()))
		return &models.Post{}, err
	}

	e := post.QueryRelatedUsers(GetSelectFieldsFromRequestFields(UserSimpleFields(), GetRequestFields(ctx)))
	if e != nil {
		return &post, e
	}

	return &post, nil
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

func (r *Resolver) Message() MessageResolver {
	return &messageResolver{r}
}

type messageResolver struct{ *Resolver }

func (r *messageResolver) Sender(ctx context.Context, obj *models.Message) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}

	sender := models.User{}
	if err := models.DB.Find(&sender, obj.SentByID); err != nil {
		err = fmt.Errorf("error finding message sentBy user with id %v ... %v", obj.SentByID, err)
		return nil, err
	}

	return &sender, nil
}

func (r *messageResolver) CreatedAt(ctx context.Context, obj *models.Message) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *messageResolver) UpdatedAt(ctx context.Context, obj *models.Message) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}

func (r *queryResolver) Message(ctx context.Context, id *string) (*models.Message, error) {
	message := models.Message{}
	messageFields := GetSelectFieldsFromRequestFields(MessageSimpleFields(), graphql.CollectAllFields(ctx))

	if err := models.DB.Select(messageFields...).Where("uuid = ?", id).First(&message); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("error getting message: %v", err.Error()))
		return &models.Message{}, err
	}

	return &message, nil
}
