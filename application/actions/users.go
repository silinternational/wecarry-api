package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/api"
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

	output, err := convertUserToAPIType(c, user)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(http.StatusOK, r.JSON(output))
}

func convertUserToAPIType(c buffalo.Context, user models.User) (api.User, error) {
	tx := models.Tx(c)

	output := api.User{}
	if err := api.ConvertToOtherType(user, &output); err != nil {
		return api.User{}, err
	}
	output.ID = user.UUID

	photoURL, err := user.GetPhotoURL(tx)
	if err != nil {
		return api.User{}, err
	}

	if photoURL != nil {
		output.AvatarURL = nulls.NewString(*photoURL)
	}

	if user.FileID.Valid {
		// depends on the earlier call to GetPhotoURL to hydrate PhotoFile
		output.PhotoID = user.PhotoFile.UUID
	}

	organizations, err := user.GetOrganizations(tx)
	if err != nil {
		return api.User{}, err
	}
	output.Organizations, err = convertOrganizationsToAPIType(organizations)
	if err != nil {
		return api.User{}, err
	}
	return output, nil
}
