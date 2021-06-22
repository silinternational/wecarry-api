package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// swagger:operation GET /users/me Users UsersMe
//
// gets the data for authenticated User.
//
// ---
// responses:
//   '200':
//     description: authenticated user
//     schema:
//       "$ref": "#/definitions/User"
func usersMe(c buffalo.Context) error {
	user := models.CurrentUser(c)

	tx := models.Tx(c)

	if err := tx.Load(&user); err != nil {
		return domain.ReportError(c, err, "GetUser")
	}

	photoURL, err := user.GetPhotoURL(tx)
	if err != nil {
		return domain.ReportError(c, err, "GetUserPhotoURL")
	}
	if photoURL != nil {
		user.AvatarURL = nulls.NewString(*photoURL)
	}

	if user.FileID.Valid {
		// depends on the earlier call to GetPhotoURL to hydrate PhotoFile
		user.PhotoID = user.PhotoFile.UUID
	}

	organizations, err := user.GetOrganizations(tx)
	if err != nil {
		return domain.ReportError(c, err, "GetUserOrganizations")
	}
	user.Organizations = organizations

	return c.Render(http.StatusOK, sheriffRenderer{value: user, groups: []string{"api"}})
}
