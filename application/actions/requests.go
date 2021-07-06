package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/cache"
	"github.com/silinternational/wecarry-api/models"
)

// swagger:operation GET /requests Requests
//
// gets the list of requests for the current user
//
// ---
// responses:
//   '200':
//     description: requests list for the current user
//     schema:
//       "$ref": "#/definitions/Requests"
func requestsList(c buffalo.Context) error {
	cUser := models.CurrentUser(c)
	tx := models.Tx(c)

	orgs, err := cUser.GetOrganizations(tx)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	requestsList, err := cache.GetVisibleRequests(c, orgs)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(200, render.JSON(requestsList))
}
