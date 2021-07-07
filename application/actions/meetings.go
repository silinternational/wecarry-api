package actions

import (
	"context"
	"errors"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
	"time"
)

// swagger:operation GET /events Events ListEvents
//
// gets a list of meetings
//
// ---
// responses:
//   '200':
//     description: meetings list
//     schema:
//       "$ref": "#/definitions/Meetings"
func meetingsList(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)
	meetings := models.Meetings{}
	if err := meetings.FindOnOrAfterDate(tx, time.Now().UTC()); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorMeetingsGet, api.CategoryInternal))
	}

	output, err := convertMeetings(c, meetings, cUser)
	if err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorMeetingsConvert, api.CategoryInternal))
	}

	return c.Render(200, render.JSON(output))
}

// converts list of model.Meeting into list of api.Meeting
func convertMeetings(ctx context.Context, meetings []models.Meeting, user models.User) ([]api.Meeting, error) {
	output := make([]api.Meeting, len(meetings))

	for i, m := range meetings {
		var err error
		output[i], err = convertMeeting(ctx, m, user)
		if err != nil {
			return []api.Meeting{}, err
		}
	}

	return output, nil
}

// converts a model.Meeting into api.Meeting
func convertMeeting(ctx context.Context, meeting models.Meeting, user models.User) (api.Meeting, error) {
	output := convertMeetingAbridged(meeting)

	createdBy, err := loadMeetingCreatedBy(ctx, meeting)
	if err != nil {
		return api.Meeting{}, err
	}
	output.CreatedBy = createdBy

	imageFile, err := loadMeetingImageFile(ctx, meeting)
	if err != nil {
		return api.Meeting{}, err
	}
	output.ImageFile = imageFile

	location, err := loadMeetingLocation(ctx, meeting)
	if err != nil {
		return api.Meeting{}, err
	}
	output.Location = location

	participants, err := loadMeetingParticipants(ctx, meeting, user)
	if err != nil {
		return api.Meeting{}, err
	}
	output.Participants = participants

	return output, nil
}

func loadMeetingCreatedBy(ctx context.Context, meeting models.Meeting) (api.User, error) {
	createdBy, err := meeting.GetCreator(models.Tx(ctx))
	if err != nil {
		return api.User{}, errors.New("loading meeting creator, " + err.Error())
	}

	outputCreatedBy, err := convertUser(ctx, *createdBy)
	if err != nil {
		err = errors.New("error converting meeting created_by user: " + err.Error())
		return api.User{}, err
	}
	return outputCreatedBy, nil
}

func loadMeetingImageFile(ctx context.Context, meeting models.Meeting) (*api.File, error) {
	imageFile, err := meeting.ImageFile(models.Tx(ctx))
	if err != nil {
		err = errors.New("error converting meeting image file: " + err.Error())
		return nil, err
	}

	if imageFile == nil {
		return nil, nil
	}

	var outputImage api.File
	if err := api.ConvertToOtherType(imageFile, &outputImage); err != nil {
		err = errors.New("error converting meeting image file to api.File: " + err.Error())
		return nil, err
	}
	outputImage.ID = imageFile.UUID
	return &outputImage, nil
}

func loadMeetingLocation(ctx context.Context, meeting models.Meeting) (*api.Location, error) {
	location, err := meeting.GetLocation(models.Tx(ctx))
	if err != nil {
		err = errors.New("error converting meeting location: " + err.Error())
		return nil, err
	}

	apiLocation := convertLocation(location)
	return &apiLocation, nil
}

func loadMeetingParticipants(ctx context.Context, meeting models.Meeting, user models.User) (api.MeetingParticipants, error) {
	tx := models.Tx(ctx)

	participants, err := meeting.Participants(tx, user)
	if err != nil {
		err = errors.New("error converting meeting participants: " + err.Error())
		return nil, err
	}

	outputParticipants, err := convertMeetingParticipants(ctx, participants)
	if err != nil {
		return nil, err
	}

	return outputParticipants, nil
}

func convertMeetingParticipants(ctx context.Context, participants models.MeetingParticipants) (api.MeetingParticipants, error) {
	output := make(api.MeetingParticipants, len(participants))
	for i := range output {
		var err error
		output[i], err = convertMeetingParticipant(ctx, participants[i])
		if err != nil {
			return output, err
		}
	}
	return output, nil
}

func convertMeetingParticipant(ctx context.Context, participant models.MeetingParticipant) (api.MeetingParticipant, error) {
	tx := models.Tx(ctx)

	output := api.MeetingParticipant{}

	user, err := participant.User(tx)
	if err != nil {
		return api.MeetingParticipant{}, err
	}

	outputUser, err := convertUser(ctx, user)
	if err != nil {
		return api.MeetingParticipant{}, err
	}
	output.User = outputUser

	output.IsOrganizer = participant.IsOrganizer

	return output, nil
}

func convertMeetingAbridged(meeting models.Meeting) api.Meeting {
	return api.Meeting{
		ID:          meeting.UUID,
		Name:        meeting.Name,
		Description: meeting.Description.String,
		StartDate:   meeting.StartDate,
		EndDate:     meeting.EndDate,
		CreatedAt:   meeting.CreatedAt,
		UpdatedAt:   meeting.UpdatedAt,
		MoreInfoURL: meeting.MoreInfoURL.String,
	}
}
