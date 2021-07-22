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
//   - name: MeetingInput
//     in: body
//     required: true
//     description: input object
//     schema:
//       "$ref": "#/definitions/MeetingInput"
//
// responses:
//   '200':
//     description: the created event/meeting
//     schema:
//       "$ref": "#/definitions/Meeting"
func meetingsCreate(c buffalo.Context) error {
	cUser := models.CurrentUser(c)

	var input api.MeetingInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal data into MeetingInput, error: " + err.Error())
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

	if err = meeting.CreateInvites(c, input.Emails); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorCreateMeeting, api.CategoryUser))
	}

	output, err := models.ConvertMeeting(c, meeting, cUser)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// swagger:operation PUT /events/{event_id} Events EventsUpdate
//
// update an event/meeting
//
// ---
// parameters:
//   - name: MeetingInput
//     in: body
//     required: true
//     description: input object
//     schema:
//       "$ref": "#/definitions/MeetingInput"
//
// responses:
//   '200':
//     description: the event/meeting
//     schema:
//       "$ref": "#/definitions/Meeting"
func meetingsUpdate(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	var input api.MeetingInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal data into MeetingInput, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	tx := models.Tx(c)

	id, err := getUUIDFromParam(c, "event_id")
	if err != nil {
		return reportError(c, err)
	}

	meeting, err := convertMeetingUpdateInput(c, input, id.String())
	if err != nil {
		return reportError(c, err)
	}

	if err = meeting.Update(tx); err != nil {
		appError := api.NewAppError(err, api.ErrorUpdateMeeting, api.CategoryUser)
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return reportError(c, appError)
	}

	output, err := models.ConvertMeeting(c, meeting, cUser)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

// convertMeetingCreateInput creates a new `Meeting` from a `MeetingInput`.
// All properties that are not `nil` are set to the value provided
func convertMeetingCreateInput(ctx context.Context, input api.MeetingInput) (models.Meeting, error) {
	tx := models.Tx(ctx)

	meeting := models.Meeting{}
	if err := parseMeetingDates(input, &meeting); err != nil {
		return models.Meeting{}, err
	}

	meeting.CreatedByID = models.CurrentUser(ctx).ID
	meeting.Name = input.Name
	meeting.Description = input.Description
	meeting.MoreInfoURL = input.MoreInfoURL

	if input.ImageFileID.Valid {
		if _, err := meeting.SetImageFile(tx, input.ImageFileID.UUID.String()); err != nil {
			err = errors.New("meeting image file ID not found, " + err.Error())
			appErr := api.NewAppError(err, api.ErrorMeetingImageIDNotFound, api.CategoryUser)
			if domain.IsOtherThanNoRows(err) {
				appErr.Category = api.CategoryDatabase
			}
			return meeting, appErr
		}
	}

	location := models.ConvertLocationInput(input.Location)
	if err := meeting.SetLocation(tx, location); err != nil {
		return meeting, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryUser)
	}

	return meeting, nil
}

// convertMeetingUpdateInput returns a `Meeting` from an existing record which has been
//  updated based on a `MeetingInput`.
// All properties in the original meeting will be overwritten.
func convertMeetingUpdateInput(ctx context.Context, input api.MeetingInput, id string) (models.Meeting, error) {
	tx := models.Tx(ctx)

	var meeting models.Meeting
	if err := meeting.FindByUUID(tx, id); err != nil {
		appError := api.NewAppError(err, api.ErrorUpdateMeeting, api.CategoryNotFound)
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return meeting, appError
	}

	if err := parseMeetingDates(input, &meeting); err != nil {
		return models.Meeting{}, err
	}

	if input.ImageFileID.Valid {
		if _, err := meeting.SetImageFile(tx, input.ImageFileID.UUID.String()); err != nil {
			err = errors.New("request photo file ID not found, " + err.Error())
			appErr := api.NewAppError(err, api.ErrorMeetingImageIDNotFound, api.CategoryUser)
			if domain.IsOtherThanNoRows(err) {
				appErr.Category = api.CategoryDatabase
			}
			return meeting, appErr
		}
	} else {
		if err := meeting.RemoveFile(tx); err != nil {
			return meeting, err
		}
	}

	if err := tx.Load(&meeting, "Location"); err != nil {
		appError := api.NewAppError(err, api.ErrorUpdateMeeting, api.CategoryInternal)
		return meeting, appError
	}

	location := models.ConvertLocationInput(input.Location)
	if location != meeting.Location {
		if err := meeting.SetLocation(tx, location); err != nil {
			return meeting, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryUser)
		}
	}

	meeting.Name = input.Name
	meeting.Description = input.Description
	meeting.MoreInfoURL = input.MoreInfoURL

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

func parseMeetingDates(input api.MeetingInput, modelMtg *models.Meeting) error {
	startDate, err := time.Parse(domain.DateFormat, input.StartDate)
	if err != nil {
		err = errors.New("failed to parse StartDate, " + err.Error())
		appErr := api.NewAppError(err, api.ErrorMeetingInvalidStartDate, api.CategoryUser)
		return appErr
	}
	modelMtg.StartDate = startDate

	endDate, err := time.Parse(domain.DateFormat, input.EndDate)
	if err != nil {
		err = errors.New("failed to parse EndDate, " + err.Error())
		appErr := api.NewAppError(err, api.ErrorMeetingInvalidEndDate, api.CategoryUser)
		return appErr
	}
	modelMtg.EndDate = endDate

	return nil
}
