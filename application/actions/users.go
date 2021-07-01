package actions

import (
	"context"
	"errors"
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
//       "$ref": "#/definitions/UserPrivate"
func usersMe(c buffalo.Context) error {
	user := models.CurrentUser(c)

	output, err := convertUserPrivate(c, user)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(http.StatusOK, r.JSON(output))
}

// swagger:operation PUT /users/me Users UsersMeUpdate
//
// Updates the data for authenticated User.
//
// ---
// parameters:
//   - name: UsersInput
//     in: body
//     required: true
//     description: input object
//     schema:
//       "$ref": "#/definitions/UsersInput"
//
// responses:
//   '200':
//     description: authenticated user
//     schema:
//       "$ref": "#/definitions/UserPrivate"
func usersMeUpdate(c buffalo.Context) error {
	user := models.CurrentUser(c)

	var input api.UsersInput
	if err := StrictBind(c, &input); err != nil {
		err = errors.New("unable to unmarshal User data into UsersInput struct, error: " + err.Error())
		return reportError(c, api.NewAppError(err, api.ErrorInvalidRequestBody, api.CategoryUser))
	}

	if input.Nickname != nil {
		user.Nickname = *input.Nickname
	}

	tx := models.Tx(c)

	var err error
	if input.PhotoID == nil {
		err = user.RemoveFile(tx)
	} else {
		_, err = user.AttachPhoto(tx, *input.PhotoID)
	}
	if err != nil {
		return reportError(c, api.NewAppError(err, api.ErrorUserUpdatePhoto, api.CategoryInternal))
	}

	if err = user.Save(tx); err != nil {
		return reportError(c, err)
	}

	output, err := convertUserPrivate(c, user)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(http.StatusOK, r.JSON(output))
}

func convertUserPrivate(ctx context.Context, user models.User) (api.UserPrivate, error) {
	tx := models.Tx(ctx)

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
		output.PhotoID = nulls.NewUUID(user.PhotoFile.UUID)
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

func convertUser(ctx context.Context, user models.User) (api.User, error) {
	tx := models.Tx(ctx)

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

	return output, nil
}
