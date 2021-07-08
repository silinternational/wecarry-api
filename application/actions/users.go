package actions

import (
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"

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

	output, err := models.ConvertUserPrivate(c, user)
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

	output, err := models.ConvertUserPrivate(c, user)
	if err != nil {
		return reportError(c, err)
	}

	return c.Render(http.StatusOK, r.JSON(output))
}
