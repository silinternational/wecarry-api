package actions

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/cache"
	"github.com/silinternational/wecarry-api/conversions"
	"github.com/silinternational/wecarry-api/domain"
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

	// TODO REMOVE print organizations of current user
	orgs, err := cUser.GetOrganizations(tx)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}
	// var output_WIP []api.RequestAbridged

	// RequestFilterParams is currently empty because the UI is not using it
	filter := models.RequestFilterParams{}

	requests := models.Requests{}
	if err := requests.FindByUser(tx, cUser, filter); err != nil {
		return domain.ReportError(c, err, "GetRequests")
	}

	output, err := conversions.ConvertRequestsAbridged(c, requests)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	var requestsList []api.RequestAbridged
	for _, org := range orgs {
		fmt.Println(org.Name)
		if err := cache.CacheWrite(c, org.Name, output); err != nil {
			return domain.ReportError(c, err, "GetRequests")
		}

		// expect error when key is missing in cache
		if err := cache.CacheRead(c, org.Name, &requestsList); err != nil {
			fmt.Println("UH-OH", err)
		} else {
			return c.Render(200, render.JSON(requestsList))
		}

	}

	return c.Render(200, render.JSON(output))
}
