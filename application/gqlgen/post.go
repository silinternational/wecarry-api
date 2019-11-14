package gqlgen

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
)

// PostFields maps GraphQL fields to their equivalent database fields. For related types, the
// foreign key field name is provided.
func PostFields() map[string]string {
	return map[string]string{
		"id":           "uuid",
		"createdBy":    "created_by_id",
		"organization": "organization_id",
		"type":         "type",
		"title":        "title",
		"description":  "description",
		"size":         "size",
		"receiver":     "receiver_id",
		"provider":     "provider_id",
		"neededAfter":  "needed_after",
		"neededBefore": "needed_before",
		"category":     "category",
		"status":       "status",
		"createdAt":    "created_at",
		"updatedAt":    "updated_at",
		"url":          "url",
		"cost":         "cost",
		"photo":        "photo_file_id",
		"destination":  "destination_id",
		"origin":       "origin_id",
	}
}

// Post returns the post resolver. It is required by GraphQL
func (r *Resolver) Post() PostResolver {
	return &postResolver{r}
}

type postResolver struct{ *Resolver }

// ID resolves the `ID` property of the post query. It provides the UUID instead of the autoincrement ID.
func (r *postResolver) ID(ctx context.Context, obj *models.Post) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

// Type resolves the `type` property of the post query. It converts the model type to the gqlgen enum type
func (r *postResolver) Type(ctx context.Context, obj *models.Post) (PostType, error) {
	if obj == nil {
		return "", nil
	}
	return PostType(obj.Type), nil
}

// CreatedBy resolves the `createdBy` property of the post query. It retrieves the related record from the database.
func (r *postResolver) CreatedBy(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}

	selectFields := GetSelectFieldsForUsers(ctx)
	creator, err := obj.GetCreator(selectFields)
	if err != nil {
		extras := map[string]interface{}{
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetPostCreator", extras)
	}

	return creator, nil
}

// Receiver resolves the `receiver` property of the post query. It retrieves the related record from the database.
func (r *postResolver) Receiver(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}

	selectFields := GetSelectFieldsForUsers(ctx)
	receiver, err := obj.GetReceiver(selectFields)
	if err != nil {
		extras := map[string]interface{}{
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetPostReceiver", extras)
	}

	return receiver, nil
}

// Provider resolves the `provider` property of the post query. It retrieves the related record from the database.
func (r *postResolver) Provider(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}

	selectFields := GetSelectFieldsForUsers(ctx)
	provider, err := obj.GetProvider(selectFields)
	if err != nil {
		extras := map[string]interface{}{
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetPostProvider", extras)
	}

	return provider, nil
}

// Organization resolves the `organization` property of the post query. It retrieves the related record from the
// database.
func (r *postResolver) Organization(ctx context.Context, obj *models.Post) (*models.Organization, error) {
	if obj == nil {
		return nil, nil
	}

	selectFields := getSelectFieldsForOrganizations(ctx)
	organization, err := obj.GetOrganization(selectFields)
	if err != nil {
		extras := map[string]interface{}{
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetPostOrganization", extras)
	}

	return organization, nil
}

// Description resolves the `description` property, converting a nulls.String to a *string.
func (r *postResolver) Description(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.Description), nil
}

// Destination resolves the `destination` property of the post query, retrieving the related record from the database.
func (r *postResolver) Destination(ctx context.Context, obj *models.Post) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	destination, err := obj.GetDestination()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostDestination")
	}

	return destination, nil
}

// Origin resolves the `origin` property of the post query, retrieving the related record from the database.
func (r *postResolver) Origin(ctx context.Context, obj *models.Post) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	origin, err := obj.GetOrigin()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostOrigin")
	}

	return origin, nil
}

// Size resolves the `size` property of the post query. It converts the model type to the gqlgen enum type
func (r *postResolver) Size(ctx context.Context, obj *models.Post) (PostSize, error) {
	if obj == nil {
		return "", nil
	}
	return PostSize(obj.Size), nil
}

// NeededAfter resolves the `neededAfter` property of the post query, converting a time.Time to a RFC3339 *string
func (r *postResolver) NeededAfter(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.NeededAfter), nil
}

// NeededBefore resolves the `neededBefore` property of the post query, converting a time.Time to a RFC3339 *string
func (r *postResolver) NeededBefore(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.NeededBefore), nil
}

// Threads resolves the `threads` property of the post query, retrieving the related records from the database.
func (r *postResolver) Threads(ctx context.Context, obj *models.Post) ([]models.Thread, error) {
	if obj == nil {
		return nil, nil
	}

	selectFields := GetSelectFieldsFromRequestFields(ThreadFields(), graphql.CollectAllFields(ctx))
	user := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	threads, err := obj.GetThreads(selectFields, user)
	if err != nil {
		extras := map[string]interface{}{
			"user":   user.Uuid,
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetPostThreads", extras)
	}

	return threads, nil
}

// URL resolves the `url` property of the post query, converting nulls.String to a *string
func (r *postResolver) URL(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.URL), nil
}

// Cost resolves the `cost` property of the post query, converting float64 to *string
func (r *postResolver) Cost(ctx context.Context, obj *models.Post) (*string, error) {
	if (obj == nil) || (!obj.Cost.Valid) {
		return nil, nil
	}

	c := strconv.FormatFloat(obj.Cost.Float64, 'f', -1, 64)
	return &c, nil
}

// Photo retrieves the file attached as the primary photo
func (r *postResolver) Photo(ctx context.Context, obj *models.Post) (*models.File, error) {
	if obj == nil {
		return nil, nil
	}

	photo, err := obj.GetPhoto()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostPhoto")
	}

	return photo, nil
}

// Files retrieves the list of files attached to the post, not including the primary photo
func (r *postResolver) Files(ctx context.Context, obj *models.Post) ([]models.File, error) {
	if obj == nil {
		return nil, nil
	}
	files, err := obj.GetFiles()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostFiles")
	}

	return files, nil
}

// IsEditable indicates whether the user is allowed to edit the post
func (r *postResolver) IsEditable(ctx context.Context, obj *models.Post) (bool, error) {
	if obj == nil {
		return false, nil
	}
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	return obj.IsEditable(cUser)
}

// Posts resolves the `posts` query
func (r *queryResolver) Posts(ctx context.Context) ([]models.Post, error) {
	posts := models.Posts{}
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	selectFields := getSelectFieldsForPosts(ctx)
	if err := posts.FindByUser(ctx, cUser, selectFields...); err != nil {
		extras := map[string]interface{}{
			"user":   cUser.Uuid,
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetPosts", extras)
	}

	return posts, nil
}

// Post resolves the `post` query
func (r *queryResolver) Post(ctx context.Context, id *string) (*models.Post, error) {
	if id == nil {
		return nil, nil
	}
	var post models.Post
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	selectFields := getSelectFieldsForPosts(ctx)
	if err := post.FindByUserAndUUID(ctx, cUser, *id, selectFields...); err != nil {
		extras := map[string]interface{}{
			"user":   cUser.Uuid,
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetPost", extras)
	}

	return &post, nil
}

// convertGqlPostInputToDBPost takes a `PostInput` and either finds a record matching the UUID given in `input.ID` or
// creates a new `models.Post` with a new UUID. In either case, all properties that are not `nil` are set to the value
// provided in `input`
func convertGqlPostInputToDBPost(ctx context.Context, input postInput, currentUser models.User) (models.Post, error) {
	post := models.Post{}

	if input.ID != nil {
		if err := post.FindByUUID(*input.ID); err != nil {
			return post, err
		}
	} else {
		if err := post.NewWithUser(input.Type.String(), currentUser); err != nil {
			return post, err
		}
	}

	if input.Status != nil {
		post.SetProviderWithStatus(input.Status.String(), currentUser)
	}

	if input.OrgID != nil {
		var org models.Organization
		err := org.FindByUUID(*input.OrgID)
		if err != nil {
			return models.Post{}, err
		}
		post.OrganizationID = org.ID
	}

	if input.Type != nil {
		post.Type = input.Type.String()
	}

	setOptionalStringField(input.Title, &post.Title)

	if input.Description != nil {
		post.Description = nulls.NewString(*input.Description)
	}

	if input.Size != nil {
		post.Size = (*input.Size).String()
	}

	if input.NeededAfter != nil {
		neededAfter, err := domain.ConvertStringPtrToDate(input.NeededAfter)
		if err != nil {
			err = fmt.Errorf("error converting NeededAfter %v ... %v", input.NeededAfter, err.Error())
			return models.Post{}, err
		}
		post.NeededAfter = neededAfter
	}

	if input.NeededBefore != nil {
		neededBefore, err := domain.ConvertStringPtrToDate(input.NeededBefore)
		if err != nil {
			err = fmt.Errorf("error converting NeededBefore %v ... %v", input.NeededBefore, err.Error())
			return models.Post{}, err
		}
		post.NeededBefore = neededBefore
	}

	setOptionalStringField(input.Category, &post.Category)

	if input.URL != nil {
		post.URL = nulls.NewString(*input.URL)
	}

	if input.Cost != nil && *(input.Cost) != "" {
		c, err := strconv.ParseFloat(*input.Cost, 64)
		if err != nil {
			err = fmt.Errorf("error converting cost %v ... %v", input.Cost, err.Error())
			return models.Post{}, err
		}
		post.Cost = nulls.NewFloat64(c)
	}

	if input.PhotoID != nil {
		if file, err := post.AttachPhoto(*input.PhotoID); err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error attaching photo to Post, %s", err.Error()))
		} else {
			post.PhotoFile = file
		}
	}

	return post, nil
}

func getSelectFieldsForPosts(ctx context.Context) []string {
	requestFields := graphql.CollectAllFields(ctx)
	selectFields := GetSelectFieldsFromRequestFields(PostFields(), requestFields)
	selectFields = append(selectFields, "id")
	return selectFields
}

type postInput struct {
	ID           *string
	Status       *PostStatus
	OrgID        *string
	Type         *PostType
	Title        *string
	Description  *string
	Destination  *LocationInput
	Origin       *LocationInput
	Size         *PostSize
	NeededAfter  *string
	NeededBefore *string
	Category     *string
	URL          *string
	Cost         *string
	PhotoID      *string
}

// CreatePost resolves the `createPost` mutation.
func (r *mutationResolver) CreatePost(ctx context.Context, input postInput) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user": cUser.Uuid,
	}

	post, err := convertGqlPostInputToDBPost(ctx, input, cUser)
	if err != nil {
		return nil, reportError(ctx, err, "CreatePost.ProcessInput", extras)
	}

	dest := convertGqlLocationInputToDBLocation(*input.Destination)
	if err1 := dest.Create(); err1 != nil {
		return nil, reportError(ctx, err1, "CreatePost.SetDestination", extras)
	}
	post.DestinationID = dest.ID

	if err2 := post.Create(); err2 != nil {
		return nil, reportError(ctx, err2, "CreatePost", extras)
	}

	if input.Origin != nil {
		if err4 := post.SetOrigin(convertGqlLocationInputToDBLocation(*input.Origin)); err4 != nil {
			return nil, reportError(ctx, err4, "CreatePost.SetOrigin", extras)
		}
	}

	return &post, nil
}

// UpdatePost resolves the `updatePost` mutation.
func (r *mutationResolver) UpdatePost(ctx context.Context, input postInput) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user": cUser.Uuid,
	}

	post, err := convertGqlPostInputToDBPost(ctx, input, cUser)
	if err != nil {
		return nil, reportError(ctx, err, "UpdatePost.ProcessInput", extras)
	}

	var dbPost models.Post
	_ = dbPost.FindByID(post.ID)
	if editable, err2 := dbPost.IsEditable(cUser); err2 != nil {
		return nil, reportError(ctx, err2, "UpdatePost.GetEditable", extras)
	} else if !editable {
		return nil, reportError(ctx, errors.New("attempt to update a non-editable post"),
			"UpdatePost.NotEditable", extras)
	}

	if err3 := post.Update(); err3 != nil {
		return nil, reportError(ctx, err3, "UpdatePost", extras)
	}

	if input.Destination != nil {
		if err4 := post.SetDestination(convertGqlLocationInputToDBLocation(*input.Destination)); err4 != nil {
			return nil, reportError(ctx, err4, "UpdatePost.SetDestination", extras)
		}
	}

	if input.Origin != nil {
		if err5 := post.SetOrigin(convertGqlLocationInputToDBLocation(*input.Origin)); err5 != nil {
			return nil, reportError(ctx, err5, "UpdatePost.SetOrigin", extras)
		}
	}

	return &post, nil
}
