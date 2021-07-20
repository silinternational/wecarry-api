package actions

import (
	"context"
	"errors"
	"time"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
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

	output, err := models.ConvertMeetings(c, meetings, cUser)
	if err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorMeetingsConvert, api.CategoryInternal))
	}

	return c.Render(200, render.JSON(output))
}

// swagger:operation POST /events Events EventsCreate
//
// create a new event/meeting
//
// ---
// parameters:
//   - name: MeetingCreateInput
//     in: body
//     required: true
//     description: input object
//     schema:
//       "$ref": "#/definitions/MeetingCreateInput"
//
// responses:
//   '200':
//     description: the created event/meeting
//     schema:
//       "$ref": "#/definitions/Meeting"
func meetingsCreate(c buffalo.Context) error {
	cUser := models.CurrentUser(c)

	var input api.MeetingCreateInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal data into MeetingCreateInput, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	tx := models.Tx(c)

	meeting, err := convertMeetingCreateInput(c, input)
	if err != nil {
		return reportError(c, err)
	}

	if err = meeting.Create(tx); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorCreateMeeting, api.CategoryUser))
	}

	output, err := models.ConvertMeeting(c, meeting, cUser)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// convertMeetingCreateInput creates a new `Meeting` from a `MeetingCreateInput`.
// All properties that are not `nil` are set to the value provided
func convertMeetingCreateInput(ctx context.Context, input api.MeetingCreateInput) (models.Meeting, error) {
	tx := models.Tx(ctx)

	meeting := models.Meeting{
		CreatedByID: models.CurrentUser(ctx).ID,
		Name:        input.Name,
		Description: input.Description,
		MoreInfoURL: input.MoreInfoURL,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	}

	if input.ImageFileID.Valid {
		if _, err := meeting.SetImageFile(tx, input.ImageFileID.UUID.String()); err != nil {
			err = errors.New("meeting image file ID not found, " + err.Error())
			appErr := api.NewAppError(err, api.ErrorCreateMeetingImageIDNotFound, api.CategoryUser)
			if domain.IsOtherThanNoRows(err) {
				appErr.Category = api.CategoryDatabase
			}
			return meeting, appErr
		}
	}

	location := models.ConvertLocationInput(input.Location)
	if err := location.Create(tx); err != nil {
		return meeting, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryUser)
	}
	meeting.LocationID = location.ID

	return meeting, nil
}

// swagger:operation POST /events/join Events JoinEvent
//
// User joins an event/meeting, which saves a new Meeting Participant object
//
// ---
// parameters:
//   - name: meeting_participant
//     in: body
//     description: input object for the new meeting participant
//     required: true
//     schema:
//       "$ref": "#/definitions/MeetingParticipantInput"
// responses:
//   '200':
//     description: a meeting
//     schema:
//       "$ref": "#/definitions/Meeting"
func meetingsJoin(c buffalo.Context) error {
	var input api.MeetingParticipantInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal MeetingParticipant data into MeetingParticipantInput struct, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	user := models.CurrentUser(c)
	tx := models.Tx(c)

	domain.NewExtra(c, "meetingID", input.MeetingID)

	meeting := models.Meeting{}
	if err := meeting.FindByUUID(tx, input.MeetingID); err != nil {
		appError := api.NewAppError(err, api.ErrorMeetingsGet, api.CategoryNotFound)
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return reportError(c, appError)
	}

	var mtgParticipant models.MeetingParticipant

	if appErr := mtgParticipant.FindOrCreate(tx, meeting, user, input.Code); appErr != nil {
		return reportError(c, appErr)
	}

	output, err := models.ConvertMeeting(c, meeting, user)
	if err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorMeetingsConvert, api.CategoryInternal))
	}

	return c.Render(200, render.JSON(output))
}
