package gqlgen

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
)

func PostFields() map[string]string {
	return map[string]string{
		"id":           "uuid",
		"createdBy":    "created_by_id",
		"organization": "organization_id",
		"type":         "type",
		"title":        "title",
		"description":  "description",
		"destination":  "destination",
		"origin":       "origin",
		"size":         "size",
		"receiver":     "receiver_id",
		"provider":     "provider_id",
		"neededAfter":  "needed_after",
		"neededBefore": "needed_before",
		"category":     "category",
		"status":       "status",
		"createdAt":    "created_at",
		"updatedAt":    "updated_at",
	}
}

func (r *Resolver) Post() PostResolver {
	return &postResolver{r}
}

type postResolver struct{ *Resolver }

func (r *postResolver) ID(ctx context.Context, obj *models.Post) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

func (r *postResolver) Type(ctx context.Context, obj *models.Post) (PostType, error) {
	if obj == nil {
		return "", nil
	}
	return PostType(obj.Type), nil
}

func (r *postResolver) CreatedBy(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetCreator(GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx)))
}

func (r *postResolver) Receiver(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetReceiver(GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx)))
}

func (r *postResolver) Provider(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetProvider(GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx)))
}

func (r *postResolver) Organization(ctx context.Context, obj *models.Post) (*models.Organization, error) {
	if obj == nil {
		return nil, nil
	}
	selectFields := GetSelectFieldsFromRequestFields(OrganizationFields(), graphql.CollectAllFields(ctx))
	return obj.GetOrganization(selectFields)
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
	if obj == nil {
		return nil, nil
	}
	selectFields := GetSelectFieldsFromRequestFields(ThreadFields(), graphql.CollectAllFields(ctx))
	return obj.GetThreads(selectFields)
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
	return obj.GetThreadIdForUser(models.GetCurrentUserFromGqlContext(ctx, TestUser))
}

func (r *queryResolver) Posts(ctx context.Context) ([]*models.Post, error) {
	var posts []*models.Post
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	selectFields := getSelectFieldsForPosts(ctx)
	if err := models.DB.Select(selectFields...).Scope(scopeUserOrgs(cUser)).All(&posts); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting posts: %v", err.Error()))
		return []*models.Post{}, err
	}

	return posts, nil
}

func (r *queryResolver) Post(ctx context.Context, id *string) (*models.Post, error) {
	post := models.Post{}
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	selectFields := getSelectFieldsForPosts(ctx)
	if err := models.DB.Select(selectFields...).Scope(scopeUserOrgs(cUser)).Where("uuid = ?", id).First(&post); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting post: %v", err.Error()))
		return &models.Post{}, err
	}

	return &post, nil
}

// ConvertGqlNewPostToDBPost does what its name says, but also ...
func ConvertGqlNewPostToDBPost(gqlPost NewPost, createdByUser models.User) (models.Post, error) {
	org, err := models.FindOrgByUUID(gqlPost.OrgID)
	if err != nil {
		return models.Post{}, err
	}

	dbPost := models.Post{}
	dbPost.Uuid = domain.GetUuid()

	dbPost.CreatedByID = createdByUser.ID
	dbPost.OrganizationID = org.ID
	dbPost.Type = gqlPost.Type.String()
	dbPost.Title = gqlPost.Title

	dbPost.Description = models.ConvertStringPtrToNullsString(gqlPost.Description)
	dbPost.Destination = models.ConvertStringPtrToNullsString(gqlPost.Destination)
	dbPost.Origin = models.ConvertStringPtrToNullsString(gqlPost.Origin)

	dbPost.Size = gqlPost.Size

	neededAfter, err := domain.ConvertStringPtrToDate(gqlPost.NeededAfter)
	if err != nil {
		err = fmt.Errorf("error converting NeededAfter %v ... %v", gqlPost.NeededAfter, err.Error())
		return models.Post{}, err
	}

	dbPost.NeededAfter = neededAfter

	neededBefore, err := domain.ConvertStringPtrToDate(gqlPost.NeededBefore)
	if err != nil {
		err = fmt.Errorf("error converting NeededBefore %v ... %v", gqlPost.NeededBefore, err.Error())
		return models.Post{}, err
	}

	dbPost.NeededBefore = neededBefore
	dbPost.Category = domain.ConvertStrPtrToString(gqlPost.Category)

	return dbPost, nil
}

func getSelectFieldsForPosts(ctx context.Context) []string {
	selectFields := GetSelectFieldsFromRequestFields(PostFields(), graphql.CollectAllFields(ctx))
	selectFields = append(selectFields, "id")
	return selectFields
}

// scope query to only include organizations for current user
func scopeUserOrgs(cUser models.User) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		orgs := cUser.GetOrgIDs()

		// convert []int to []interface{}
		s := make([]interface{}, len(orgs))
		for i, v := range orgs {
			s[i] = v
		}

		return q.Where("organization_id IN (?)", s...)
	}
}
