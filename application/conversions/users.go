package conversions

import (
	"context"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func ConvertUserPrivate(ctx context.Context, user models.User) (api.UserPrivate, error) {
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

func ConvertUser(ctx context.Context, user models.User) (api.User, error) {
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
