package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/models"
)

// Meeting returns the meeting resolver. It is required by GraphQL
func (r *Resolver) Meeting() MeetingResolver {
	return &meetingResolver{r}
}

type meetingResolver struct{ *Resolver }

// ID resolves the `ID` property of the meeting query. It provides the UUID instead of the autoincrement ID.
func (r *meetingResolver) ID(ctx context.Context, obj *models.Meeting) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// CreatedBy resolves the `createdBy` property of the meeting query. It retrieves the related record from the database.
func (r *meetingResolver) CreatedBy(ctx context.Context, obj *models.Meeting) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	creator, err := obj.GetCreator()
	if err != nil {
		return nil, reportError(ctx, err, "GetMeetingCreator")
	}

	return getPublicProfile(ctx, creator), nil
}

// Description resolves the `description` property, converting a nulls.String to a *string.
func (r *meetingResolver) Description(ctx context.Context, obj *models.Meeting) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.Description), nil
}

// Location resolves the `location` property of the meeting query, retrieving the related record from the database.
func (r *meetingResolver) Location(ctx context.Context, obj *models.Meeting) (*models.Location, error) {
	if obj == nil {
		return &models.Location{}, nil
	}

	location, err := obj.GetLocation()
	if err != nil {
		return &models.Location{}, reportError(ctx, err, "GetMeetingLocation")
	}

	return &location, nil
}

// Description resolves the `description` property, converting a nulls.String to a *string.
func (r *meetingResolver) MoreInfoURL(ctx context.Context, obj *models.Meeting) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.MoreInfoURL), nil
}

// Image retrieves the file associated with the meeting
func (r *meetingResolver) Image(ctx context.Context, obj *models.Meeting) (*models.File, error) {
	if obj == nil {
		return nil, nil
	}

	image, err := obj.GetImage()
	if err != nil {
		return nil, reportError(ctx, err, "GetMeetingImage")
	}

	return image, nil
}

// Meetings resolves the `meetings` query by getting a list of meetings that have an
// end date in the future
func (r *queryResolver) Meetings(ctx context.Context) ([]models.Meeting, error) {
	meetings := models.Meetings{}
	if err := meetings.FindCurrentAndFuture(); err != nil {
		extras := map[string]interface{}{}
		return nil, reportError(ctx, err, "GetMeetings", extras)
	}

	return meetings, nil
}
