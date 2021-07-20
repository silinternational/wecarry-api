package actions

import (
	"errors"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

// swagger:operation POST /watches Watches CreateWatch
//
// Create a new Watch for the User
//
// ---
// parameters:
//   - name: watch
//     in: body
//     description: watch input object
//     required: true
//     schema:
//       "$ref": "#/definitions/WatchInput"
// responses:
//   '200':
//     description: {"id": "<the id of the new watch>"}
func watchesCreate(c buffalo.Context) error {
	var input api.WatchInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal Watch data into WatchInput struct, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	if input.IsEmpty() {
		err := errors.New("empty WatchInput is not allowed")
		return reportError(c, api.NewAppError(err, api.ErrorWatchInputEmpty, api.CategoryUser))
	}

	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	newWatch, err := convertWatchInput(tx, input, cUser)
	if err != nil {
		err := errors.New("unable to find Meeting related to a new Watch, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorWatchInputMeetingFailure, api.CategoryUser))
	}

	if input.Destination != nil {
		location := models.ConvertLocationInput(*input.Destination)
		if err = location.Create(tx); err != nil {
			err := errors.New("unable to create the destination related to a new Watch, error: " + err.Error())
			return reportError(c, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryInternal))
		}
		newWatch.DestinationID = nulls.NewInt(location.ID)
	}

	if input.Origin != nil {
		location := models.ConvertLocationInput(*input.Origin)
		if err = location.Create(tx); err != nil {
			err := errors.New("unable to create the origin related to a new Watch, error: " + err.Error())
			return reportError(c, api.NewAppError(err, api.ErrorLocationCreateFailure, api.CategoryInternal))
		}
		newWatch.OriginID = nulls.NewInt(location.ID)
	}

	if err = newWatch.Create(tx); err != nil {
		err := errors.New("unable to create the new Watch, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorWatchCreateFailure, api.CategoryInternal))
	}

	output := map[string]string{"id": newWatch.UUID.String()}

	return c.Render(200, render.JSON(output))
}

// swagger:operation DELETE /watches/{watch_id} Watches RemoveWatch
//
// Remove one of the User's Watches
//
// ---
// responses:
//   '200':
//     description: The id (uuid) of the deleted watch
func watchesRemove(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)
	id := getWatchIDFromContext(c)

	var watch models.Watch
	output, appErr := watch.DeleteForOwner(tx, id, cUser)
	if appErr != nil {
		return reportError(c, appErr)
	}

	return c.Render(200, render.JSON(output))
}

// swagger:operation GET /watches Watches UsersWatches
//
// List the User's Watches
//
// ---
// responses:
//   '200':
//     description: A list of the user's watches
//     schema:
//       "$ref": "#/definitions/Watches"
func watchesMine(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	watches := models.Watches{}
	if err := watches.FindByUser(tx, cUser, "Owner", "Destination", "Origin", "Meeting"); err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorWatchesLoadFailure, api.CategoryInternal))
	}

	output, err := convertWatches(tx, watches, cUser)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(200, render.JSON(output))
}

func convertWatches(tx *pop.Connection, watches models.Watches, user models.User) (api.Watches, error) {
	output := make(api.Watches, len(watches))

	for i := range output {
		next, err := convertWatch(tx, watches[i], user)
		if err != nil {
			return nil, err
		}
		output[i] = next
	}

	return output, nil
}

func convertWatch(tx *pop.Connection, watch models.Watch, user models.User) (api.Watch, error) {
	var output api.Watch
	if err := api.ConvertToOtherType(watch, &output); err != nil {
		err = errors.New("error converting watches to api.Watches: " + err.Error())
		return api.Watch{}, err
	}

	if err := watch.LoadForAPI(tx, user); err != nil {
		err = errors.New("error hydrating watch: " + err.Error())
	}

	output.ID = watch.UUID

	if !watch.MeetingID.Valid {
		output.Meeting = nil
	} else {
		output.Meeting.ID = watch.Meeting.UUID
	}

	if !watch.DestinationID.Valid {
		output.Destination = nil
	}

	if !watch.OriginID.Valid {
		output.Origin = nil
	}

	return output, nil
}

// convertWatchInput takes a `WatchInput` and either finds a record matching the UUID given in `input.ID` or
// creates a new `models.Watch` with a new UUID. In either case, all properties that are not `nil` are set to the value
// provided in `input`
func convertWatchInput(tx *pop.Connection, input api.WatchInput, user models.User) (models.Watch, error) {
	watch := models.Watch{}

	watch.OwnerID = user.ID
	watch.Name = input.Name

	watch.SearchText = input.SearchText

	if input.Size == nil {
		watch.Size = nil
	} else {
		apiSize := *input.Size
		s := models.GetRequestSizeFromAPISize(apiSize)
		watch.Size = &s
	}

	if !input.MeetingID.Valid {
		watch.MeetingID = nulls.Int{}
	} else {
		var meeting models.Meeting
		if err := meeting.FindByUUID(tx, input.MeetingID.UUID.String()); err != nil {
			return watch, err
		}
		watch.MeetingID = nulls.NewInt(meeting.ID)
	}

	return watch, nil
}
