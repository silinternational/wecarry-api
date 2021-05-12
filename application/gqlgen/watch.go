package gqlgen

import (
	"context"
	"errors"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// Watch returns the watch resolver. It is required by GraphQL
func (r *Resolver) Watch() WatchResolver {
	return &watchResolver{r}
}

type watchResolver struct{ *Resolver }

// ID resolves the `ID` property of the watch query. It provides the UUID instead of the autoincrement ID.
func (r *watchResolver) ID(ctx context.Context, obj *models.Watch) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// Owner resolves the `owner` property of the watch query. It retrieves the related record from the database.
func (r *watchResolver) Owner(ctx context.Context, obj *models.Watch) (*PublicProfile, error) {
	if obj == nil {
		return &PublicProfile{}, nil
	}

	creator, err := obj.GetOwner()
	if err != nil {
		return &PublicProfile{}, domain.ReportError(ctx, err, "GetWatchCreator")
	}

	return getPublicProfile(ctx, creator), nil
}

// Destination is a field resolver
func (r *watchResolver) Destination(ctx context.Context, obj *models.Watch) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	location, err := obj.GetDestination()
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetWatchDestination")
	}

	return location, nil
}

// Origin is a field resolver
func (r *watchResolver) Origin(ctx context.Context, obj *models.Watch) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	location, err := obj.GetOrigin()
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetWatchOrigin")
	}

	return location, nil
}

// SearchText resolves the `searchText` property of the watch query
func (r *watchResolver) SearchText(ctx context.Context, obj *models.Watch) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	if !obj.SearchText.Valid {
		return nil, nil
	}

	return &obj.SearchText.String, nil
}

// MyWatches resolves the `myWatches` query by getting a list of Watches owned by the current user
func (r *queryResolver) MyWatches(ctx context.Context) ([]models.Watch, error) {
	watches := models.Watches{}
	currentUser := models.CurrentUser(ctx)
	if err := watches.FindByUser(currentUser); err != nil {
		return nil, domain.ReportError(ctx, err, "MyWatches")
	}

	return watches, nil
}

// convertWatchInput takes a `WatchInput` and either finds a record matching the UUID given in `input.ID` or
// creates a new `models.Watch` with a new UUID. In either case, all properties that are not `nil` are set to the value
// provided in `input`
func convertWatchInput(ctx context.Context, input watchInput, currentUser models.User) (models.Watch, error) {
	watch := models.Watch{}

	if input.ID != nil {
		if err := watch.FindByUUID(*input.ID); err != nil {
			return watch, err
		}
	} else {
		watch.OwnerID = currentUser.ID
	}

	watch.Name = input.Name
	watch.SearchText = models.ConvertStringPtrToNullsString(input.SearchText)

	if input.Size == nil {
		watch.Size = nil
	} else {
		s := *input.Size
		watch.Size = &s
	}

	if input.MeetingID == nil || *input.MeetingID == "" {
		watch.MeetingID = nulls.Int{}
	} else {
		var meeting models.Meeting
		if err := meeting.FindByUUID(*input.MeetingID); err != nil {
			return watch, err
		}
		watch.MeetingID = nulls.NewInt(meeting.ID)
	}

	return watch, nil
}

type watchInput struct {
	ID          *string
	Name        string
	Destination *LocationInput
	Origin      *LocationInput
	MeetingID   *string
	SearchText  *string
	Size        *models.RequestSize
}

// CreateWatch resolves the `createWatch` mutation.
func (r *mutationResolver) CreateWatch(ctx context.Context, input watchInput) (*models.Watch, error) {
	cUser := models.CurrentUser(ctx)

	watch, err := convertWatchInput(ctx, input, cUser)
	if err != nil {
		return &models.Watch{}, domain.ReportError(ctx, err, "CreateWatch.ProcessInput")
	}

	if input.Destination != nil {
		location := convertLocation(*input.Destination)
		if err = location.Create(); err != nil {
			return &models.Watch{}, domain.ReportError(ctx, err, "CreateWatch.SetLocation")
		}
		watch.DestinationID = nulls.NewInt(location.ID)
	}

	if input.Origin != nil {
		location := convertLocation(*input.Origin)
		if err = location.Create(); err != nil {
			return nil, domain.ReportError(ctx, err, "CreateWatch.SetOrigin")
		}
		watch.OriginID = nulls.NewInt(location.ID)
	}

	if err = watch.Create(); err != nil {
		return &models.Watch{}, domain.ReportError(ctx, err, "CreateWatch")
	}

	return &watch, nil
}

// UpdateWatch resolves the `updateWatch` mutation.
func (r *mutationResolver) UpdateWatch(ctx context.Context, input watchInput) (*models.Watch, error) {
	currentUser := models.CurrentUser(ctx)
	watch, err := convertWatchInput(ctx, input, currentUser)
	if err != nil {
		return &models.Watch{}, domain.ReportError(ctx, err, "UpdateWatch.ProcessInput")
	}

	if watch.OwnerID != currentUser.ID {
		return &models.Watch{}, domain.ReportError(ctx, errors.New("user attempted to update non-owned Watch"),
			"UpdateWatch.NotFound")
	}

	if err := watch.Update(); err != nil {
		return &models.Watch{}, domain.ReportError(ctx, err, "UpdateWatch")
	}

	if input.Destination != nil {
		if err = watch.SetDestination(convertLocation(*input.Destination)); err != nil {
			return &models.Watch{}, domain.ReportError(ctx, err, "UpdateWatch.SetDestination")
		}
	}

	if input.Origin != nil {
		if err = watch.SetOrigin(convertLocation(*input.Origin)); err != nil {
			return nil, domain.ReportError(ctx, err, "UpdateWatch.SetOrigin")
		}
	}

	return &watch, nil
}

// RemoveWatch resolves the `removeWatch` mutation.
func (r *mutationResolver) RemoveWatch(ctx context.Context, input RemoveWatchInput) ([]models.Watch, error) {
	currentUser := models.CurrentUser(ctx)

	var watch models.Watch
	if err := watch.FindByUUID(input.ID); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveWatch.NotFound")
	}

	if watch.OwnerID != currentUser.ID {
		return nil, domain.ReportError(ctx, errors.New("user attempted to delete non-owned Watch"),
			"RemoveWatch.NotFound")
	}

	if err := watch.Destroy(); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveWatch")
	}

	var watches models.Watches
	if err := watches.FindByUser(currentUser); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveWatch.FindByUser")
	}

	return watches, nil
}
