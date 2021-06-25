package actions

import (
	"context"
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v5"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

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
	if err := watches.FindByUser(tx, cUser, "Owner", "Destination", "Origin"); err != nil {
		return reportError(c, &api.AppError{
			HttpStatus: http.StatusInternalServerError,
			Key:        api.WatchesLoadFailure,
			Err:        err,
		})
	}

	var output api.Watches
	output, err := convertWatchesToAPIType(c, tx, watches, cUser)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(output))
}

func convertWatchesToAPIType(ctx context.Context, tx *pop.Connection, watches models.Watches, user models.User) (api.Watches, error) {
	var output api.Watches

	if err := api.ConvertToOtherType(watches, &output); err != nil {
		err = errors.New("error converting watches to api.Watches: " + err.Error())
		return nil, err
	}

	// Hydrate the watches' own and related fields
	for i := range output {
		if err := watches[i].LoadForAPI(tx, user); err != nil {
			err = errors.New("error converting watch to api.Watch: " + err.Error())
		}

		ownerOutput, err := convertUserToAPIType(ctx, watches[i].Owner)
		if err != nil {
			err = errors.New("error converting watch meeting to api.MeetingName: " + err.Error())
			return api.Watches{}, err
		}

		output[i].Owner = ownerOutput

		if watches[i].MeetingID.Valid {
			var meetingOutput api.MeetingName
			if err := api.ConvertToOtherType(watches[i].Meeting, &meetingOutput); err != nil {
				err = errors.New("error converting watch meeting to api.MeetingName: " + err.Error())
				return api.Watches{}, err
			}
			output[i].Meeting = &meetingOutput
		}

		output[i].ID = watches[i].UUID

		if !watches[i].DestinationID.Valid {
			output[i].Destination = nil
		}

		if !watches[i].OriginID.Valid {
			output[i].Origin = nil
		}
	}

	return output, nil
}
