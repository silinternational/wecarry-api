package actions

import (
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
