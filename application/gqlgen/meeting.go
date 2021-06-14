package gqlgen

import (
	"context"
	"fmt"
	"time"

	"github.com/silinternational/wecarry-api/dataloader"
	"github.com/silinternational/wecarry-api/domain"
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
		return &PublicProfile{}, nil
	}

	creator, err := dataloader.For(ctx).UsersByID.Load(obj.CreatedByID)
	if err != nil {
		return &PublicProfile{}, domain.ReportError(ctx, err, "GetMeetingCreator")
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

	location, err := dataloader.For(ctx).LocationsByID.Load(obj.LocationID)
	if err != nil {
		return &models.Location{}, domain.ReportError(ctx, err, "GetMeetingLocation")
	}
	return location, nil
}

// MoreInfoURL resolves the `moreInfoURL` property, converting a nulls.String to a *string.
func (r *meetingResolver) MoreInfoURL(ctx context.Context, obj *models.Meeting) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.MoreInfoURL), nil
}

// StartDate resolves the `startDate` property, converting a time.Time to a string.
func (r *meetingResolver) StartDate(ctx context.Context, obj *models.Meeting) (string, error) {
	if obj == nil {
		return "", nil
	}
	date := obj.StartDate.Format(domain.DateFormat)
	return date, nil
}

// EndDate resolves the `endDate` property, converting a time.Time to a string.
func (r *meetingResolver) EndDate(ctx context.Context, obj *models.Meeting) (string, error) {
	if obj == nil {
		return "", nil
	}
	date := obj.EndDate.Format(domain.DateFormat)
	return date, nil
}

// ImageFile retrieves the file associated with the meeting
func (r *meetingResolver) ImageFile(ctx context.Context, obj *models.Meeting) (*models.File, error) {
	if obj == nil {
		return nil, nil
	}

	if !obj.FileID.Valid {
		return nil, nil
	}

	image, err := dataloader.For(ctx).FilesByID.Load(obj.FileID.Int)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetMeetingImage")
	}

	return image, nil
}

func (r *meetingResolver) Requests(ctx context.Context, obj *models.Meeting) ([]models.Request, error) {
	if obj == nil {
		return nil, nil
	}
	requests, err := obj.Requests(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "Meeting.Requests")
	}
	return requests, nil
}

func (r *meetingResolver) Invites(ctx context.Context, obj *models.Meeting) ([]models.MeetingInvite, error) {
	if obj == nil {
		return nil, nil
	}
	invites, err := obj.Invites(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "Meeting.Invites")
	}
	return invites, nil
}

func (r *meetingResolver) Participants(ctx context.Context, obj *models.Meeting) ([]models.MeetingParticipant, error) {
	if obj == nil {
		return nil, nil
	}
	participants, err := obj.Participants(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "Meeting.Participants")
	}
	return participants, nil
}

func (r *meetingResolver) Visibility(ctx context.Context, obj *models.Meeting) (MeetingVisibility, error) {
	return MeetingVisibilityAll, nil
}

func (r *meetingResolver) Organizers(ctx context.Context, obj *models.Meeting) ([]PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}
	users, err := obj.Organizers(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "Meeting.Organizers")
	}
	return getPublicProfiles(ctx, users), nil
}

// Meetings resolves the `meetings` query by getting a list of meetings that have an
// end date in the future
func (r *queryResolver) Meetings(ctx context.Context, endAfter, endBefore, startAfter, startBefore *string) ([]models.Meeting, error) {
	meetings := models.Meetings{}
	if err := meetings.FindOnOrAfterDate(ctx, time.Now()); err != nil {
		return nil, domain.ReportError(ctx, err, "GetMeetings")
	}

	return meetings, nil
}

// RecentMeetings resolves the `recentMeetings` query by getting a list of meetings that have an
// end date in the last <domain.RecentMeetingDelay> time period
func (r *queryResolver) RecentMeetings(ctx context.Context) ([]models.Meeting, error) {
	meetings := models.Meetings{}
	if err := meetings.FindRecent(ctx, time.Now()); err != nil {
		return nil, domain.ReportError(ctx, err, "GetRecentMeetings")
	}

	return meetings, nil
}

// Meeting resolves the `meeting` query
func (r *queryResolver) Meeting(ctx context.Context, id *string) (*models.Meeting, error) {
	if id == nil {
		return &models.Meeting{}, nil
	}
	var meeting models.Meeting
	if err := meeting.FindByUUID(models.Tx(ctx), *id); err != nil {
		return &models.Meeting{}, domain.ReportError(ctx, err, "GetMeeting")
	}

	return &meeting, nil
}

// convertGqlMeetingInputToDBMeeting takes a `MeetingInput` and either finds a record matching the UUID given in `input.ID` or
// creates a new `models.Meeting` with a new UUID. In either case, all properties that are not `nil` are set to the value
// provided in `input`
func convertGqlMeetingInputToDBMeeting(ctx context.Context, input meetingInput, currentUser models.User) (models.Meeting, error) {
	meeting := models.Meeting{}
	if input.ID != nil {
		if err := meeting.FindByUUID(models.Tx(ctx), *input.ID); err != nil {
			return meeting, err
		}
	} else {
		meeting.CreatedByID = currentUser.ID
	}

	setStringField(input.Name, &meeting.Name)
	meeting.Description = models.ConvertStringPtrToNullsString(input.Description)
	meeting.MoreInfoURL = models.ConvertStringPtrToNullsString(input.MoreInfoURL)

	if input.StartDate != nil {
		startTime, err := domain.ConvertStringPtrToDate(input.StartDate)
		if err != nil {
			return models.Meeting{}, err
		}
		meeting.StartDate = startTime
	}

	if input.EndDate != nil {
		endTime, err := domain.ConvertStringPtrToDate(input.EndDate)
		if err != nil {
			return models.Meeting{}, err
		}
		meeting.EndDate = endTime
	}

	var err error
	if input.ImageFileID != nil {
		_, err = meeting.SetImageFile(ctx, *input.ImageFileID)
	} else if meeting.ID > 0 {
		err = meeting.RemoveFile(ctx)
	}
	if err != nil {
		return meeting, domain.ReportError(ctx, fmt.Errorf("error updating meeting image file, %s",
			err.Error()), "Meeting.UpdateImageFile")
	}

	return meeting, nil
}

type meetingInput struct {
	ID          *string
	Name        *string
	Description *string
	StartDate   *string
	EndDate     *string
	MoreInfoURL *string
	ImageFileID *string
	Location    *LocationInput
	Visibility  MeetingVisibility
}

// CreateMeeting resolves the `createMeeting` mutation.
func (r *mutationResolver) CreateMeeting(ctx context.Context, input meetingInput) (*models.Meeting, error) {
	cUser := models.CurrentUser(ctx)

	meeting, err := convertGqlMeetingInputToDBMeeting(ctx, input, cUser)
	if err != nil {
		return &models.Meeting{}, domain.ReportError(ctx, err, "CreateMeeting.ProcessInput")
	}

	if !meeting.CanCreate(cUser) {
		return &models.Meeting{}, domain.ReportError(ctx, err, "CreateMeeting.Unauthorized")
	}

	location := convertLocation(*input.Location)
	if err = location.Create(ctx); err != nil {
		return &models.Meeting{}, domain.ReportError(ctx, err, "CreateMeeting.SetLocation")
	}
	meeting.LocationID = location.ID

	if err = meeting.Create(ctx); err != nil {
		return &models.Meeting{}, domain.ReportError(ctx, err, "CreateMeeting")
	}

	return &meeting, nil
}

// UpdateMeeting resolves the `updateMeeting` mutation.
func (r *mutationResolver) UpdateMeeting(ctx context.Context, input meetingInput) (*models.Meeting, error) {
	cUser := models.CurrentUser(ctx)
	meeting, err := convertGqlMeetingInputToDBMeeting(ctx, input, cUser)
	if err != nil {
		return &models.Meeting{}, domain.ReportError(ctx, err, "UpdateMeeting.ProcessInput")
	}

	if !meeting.CanUpdate(cUser) {
		return &models.Meeting{}, domain.ReportError(ctx, err, "UpdateMeeting.Unauthorized")
	}

	if err := meeting.Update(ctx); err != nil {
		return &models.Meeting{}, domain.ReportError(ctx, err, "UpdateMeeting")
	}

	if input.Location != nil {
		if err = meeting.SetLocation(ctx, convertLocation(*input.Location)); err != nil {
			return &models.Meeting{}, domain.ReportError(ctx, err, "UpdateMeeting.SetLocation")
		}
	}

	return &meeting, nil
}
