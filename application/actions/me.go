package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/silinternational/handcarry-api/models"
)

func MeHandler(c buffalo.Context) error {
	return c.Render(200, r.JSON(models.GetCurrentUser(c)))
}
