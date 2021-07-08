package actions

import (
	"errors"
	"github.com/silinternational/wecarry-api/domain"
	"time"

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

// swagger:operation POST /events Events JoinEvent
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
