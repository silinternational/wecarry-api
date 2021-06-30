package actions

import (
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v5"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

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

	id, err := getUUIDFromParam(c, "watch_id")
	if err != nil {
		return reportError(c, err)
	}

	var watch models.Watch
	output, appErr := watch.DeleteForOwner(tx, id.String(), cUser)
	if appErr != nil {
		appErr.SetHttpStatusFromCategory()
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
		return reportError(c, &api.AppError{
			HttpStatus: http.StatusInternalServerError,
			Key:        api.WatchesLoadFailure,
			Err:        err,
		})
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
