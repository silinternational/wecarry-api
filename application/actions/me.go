package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/silinternational/handcarry-api/domain"
)

func MeHandler(c buffalo.Context) error {
	return c.Render(200, r.JSON(domain.GetCurrentUser(c)))
}
