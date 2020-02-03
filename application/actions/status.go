package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/models"
)

// statusHandler is a handler to respond with the current site status
func statusHandler(c buffalo.Context) error {
	var orgs models.Organizations
	if err := orgs.All(); err != nil {
		c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"status": "error"}))
	}
	return c.Render(200, r.JSON(map[string]string{"status": "good"}))
}
