package actions

import (
	"context"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

// swagger:operation GET /users/me Users UsersMe
//
// gets the data for authenticated UserPrivate.
//
// ---
// responses:
//   '200':
//     description: authenticated user
//     schema:
//       "$ref": "#/definitions/UserPrivate"
func usersMe(c buffalo.Context) error {
	user := models.CurrentUser(c)

	output, err := convertUserToPrivateAPIType(c, user)
	if err != nil {
		return reportError(c, appErrorFromErr(err))
	}

	return c.Render(http.StatusOK, r.JSON(output))
}

func convertUserToPrivateAPIType(c context.Context, user models.User) (api.UserPrivate, error) {
	tx := models.Tx(c)

	output := api.UserPrivate{}
	if err := api.ConvertToOtherType(user, &output); err != nil {
		return api.UserPrivate{}, err
	}
	output.ID = user.UUID

	photoURL, err := user.GetPhotoURL(tx)
	if err != nil {
		return api.UserPrivate{}, err
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
		return api.UserPrivate{}, err
	}
	output.Organizations, err = convertOrganizationsToAPIType(organizations)
	if err != nil {
		return api.UserPrivate{}, err
	}
	return output, nil
}

func convertUserToAPIType(c context.Context, user models.User) (api.User, error) {
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

	if err != nil {
		return api.User{}, err
	}
	return output, nil
}
