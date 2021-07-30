package actions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"

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
		return reportError(c, api.NewAppError(err, api.ErrorMeetingCreate, api.CategoryUser))
	}

	if err = meeting.CreateInvites(c, input.Emails); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorMeetingCreate, api.CategoryUser))
	}

	var mtgParticipant models.MeetingParticipant
	if appErr := mtgParticipant.FindOrCreate(tx, meeting, cUser, nil); appErr != nil {
		return reportError(c, appErr)
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
	domain.NewExtra(c, "userID", cUser.ID)

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

	domain.NewExtra(c, "meetingID", meeting.ID)

	if !cUser.CanUpdateMeeting(meeting) {
		err := errors.New("user is not authorized to update meeting")
		return reportError(c, api.NewAppError(err, api.ErrorNotAuthorized, api.CategoryForbidden))
	}

	if err = meeting.Update(tx); err != nil {
		appError := api.NewAppError(err, api.ErrorMeetingUpdate, api.CategoryUser)
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return reportError(c, appError)
	}

	output, err := models.ConvertMeeting(c, meeting, cUser, models.OptIncludeParticipants)
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
		appError := api.NewAppError(err, api.ErrorMeetingUpdate, api.CategoryNotFound)
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
		appError := api.NewAppError(err, api.ErrorMeetingUpdate, api.CategoryInternal)
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

// swagger:operation GET /events/{event_id} Events GetEvent
//
// gets one event/meeting with its participants
//
// ---
// responses:
//   '200':
//     description: an event/meeting
//     schema:
//       "$ref": "#/definitions/Meeting"
func meetingsGet(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	id, err := getUUIDFromParam(c, "event_id")
	if err != nil {
		return reportError(c, err)
	}

	meeting := models.Meeting{}
	if err = meeting.FindByUUID(tx, id.String()); err != nil {
		appError := api.NewAppError(err, api.ErrorMeetingGet, api.CategoryNotFound)
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return reportError(c, appError)
	}

	canViewParticipants, err := cUser.CanViewMeetingParticipants(tx, meeting)
	if err != nil {
		appError := api.NewAppError(err, api.ErrorMeetingGet, api.CategoryInternal)
		return reportError(c, appError)
	}

	var option models.MeetingOption
	if canViewParticipants {
		option = models.OptIncludeParticipants
	}

	isDeletable, err := meeting.CanDelete(tx, cUser)
	if err != nil {
		appError := api.NewAppError(err, api.ErrorMeetingGet, api.CategoryInternal)
		return reportError(c, appError)
	}

	output, err := models.ConvertMeeting(c, meeting, cUser, option)
	if err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorMeetingsConvert, api.CategoryInternal))
	}

	output.IsDeletable = nulls.NewBool(isDeletable)

	return c.Render(200, render.JSON(output))
}

// swagger:operation DELETE /events/{event_id} Events DeleteEvent
//
// delete one event/meeting with its participants as long as there are no requests
// associated with it
//
// ---
// responses:
//   '204':
//     description: OK but no content in response
func meetingsRemove(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	id, err := getUUIDFromParam(c, "event_id")
	if err != nil {
		return reportError(c, err)
	}

	meeting := models.Meeting{}
	if err = meeting.FindByUUID(tx, id.String()); err != nil {
		appError := api.NewAppError(err, api.ErrorMeetingGet, api.CategoryNotFound)
		if domain.IsOtherThanNoRows(err) {
			appError.Category = api.CategoryInternal
		}
		return reportError(c, appError)
	}

	if !cUser.CanUpdateMeeting(meeting) {
		err := errors.New("user is not authorized to delete the meeting")
		return reportError(c, api.NewAppError(err, api.ErrorNotAuthorized, api.CategoryForbidden))
	}

	if err := meeting.SafeDelete(tx); err != nil {
		appError := api.NewAppError(err, api.ErrorMeetingDelete, api.CategoryInternal)
		if strings.Contains(err.Error(), `meeting with associated requests may not be deleted`) {
			appError.Category = api.CategoryUser
		}
		return reportError(c, appError)
	}

	return c.Render(http.StatusNoContent, nil)
}

// swagger:operation DELETE /events/{event_id}/invite Events DeleteEventInvite
//
// delete one invite from an event/meeting
//
// ---
// parameters:
//   - name: invite email
//     type: string
//     in: body
//     description: email of invite to be deleted
//     required: true
//     schema:
//       "$ref": "#/definitions/MeetingInviteEmail"
// responses:
//   '204':
//     description: OK but no content in response
func meetingsInviteDelete(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	meetingID, err := getUUIDFromParam(c, "event_id")
	if err != nil {
		return reportError(c, err)
	}

	input := api.MeetingInviteEmail{}
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal data into MeetingInviteEmail, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	inviteEmail := input.InviteEmail

	var meeting models.Meeting
	if err := meeting.FindByUUID(tx, meetingID.String()); err != nil {
		err := fmt.Errorf("could not find meeting with id: '%s'", meetingID)
		appError := api.NewAppError(err, api.ErrorMeetingInviteDelete, api.CategoryNotFound)
		return reportError(c, appError)
	}

	if !cUser.CanUpdateMeeting(meeting) {
		err := errors.New("user is not authorized to delete the meeting invite")
		return reportError(c, api.NewAppError(err, api.ErrorNotAuthorized, api.CategoryForbidden))
	}

	var invite models.MeetingInvite
	if err := invite.FindByMeetingIDAndEmail(tx, meeting.ID, inviteEmail); err != nil {
		err := fmt.Errorf("could not find meeting invite with meeting id: '%v' and invite email: %s",
			meeting.ID, inviteEmail)
		appError := api.NewAppError(err, api.ErrorMeetingInviteDelete, api.CategoryNotFound)
		return reportError(c, appError)
	}

	if err := invite.Destroy(tx); err != nil {
		appError := api.NewAppError(err, api.ErrorMeetingInviteDelete, api.CategoryInternal)
		return reportError(c, appError)
	}

	return c.Render(http.StatusNoContent, nil)
}

func meetingsInvitePost(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	meetingID, err := getUUIDFromParam(c, "event_id")
	if err != nil {
		return reportError(c, err)
	}

	input := api.MeetingInviteEmails{}
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal data into MeetingInviteEmails, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorMeetingInvitesPost, api.CategoryUser))
	}

	var meeting models.Meeting
	if err := meeting.FindByUUID(tx, meetingID.String()); err != nil {
		err := fmt.Errorf("could not find meeting with id: '%s'", meetingID)
		appError := api.NewAppError(err, api.ErrorMeetingInvitesPost, api.CategoryNotFound)
		return reportError(c, appError)
	}

	if !cUser.CanUpdateMeeting(meeting) {
		err := errors.New("user is not authorized to add meeting invites")
		return reportError(c, api.NewAppError(err, api.ErrorNotAuthorized, api.CategoryForbidden))
	}

	invite := models.MeetingInvite{
		MeetingID: meeting.ID,
		InviterID: cUser.ID,
	}
	for _, i := range input {
		err := invite.FindByMeetingIDAndEmail(tx, meeting.ID, i)
		if err == nil {
			err := fmt.Errorf("meeting invite already exists with meeting id: '%v' and invite email: %s",
				meeting.ID, i)
			appError := api.NewAppError(err, api.ErrorMeetingInvitesPost, api.CategoryUser)
			return reportError(c, appError)
		} else {
			invite.Email = i
			err := invite.Create(tx)
			if err != nil {
				return reportError(c, api.NewAppError(err, api.ErrorMeetingInvitesPost, api.CategoryUser))
			}
		}
	}

	return c.Render(http.StatusNoContent, nil)
}
