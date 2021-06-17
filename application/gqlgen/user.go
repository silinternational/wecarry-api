package gqlgen

import (
	"context"
	"errors"
	"strings"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// requestRoleMap is used to convert RequestRole gql enum values to values used by models
var requestRoleMap = map[RequestRole]string{
	RequestRoleCreatedby: models.RequestsCreated,
	RequestRoleProviding: models.RequestsProviding,
}

// User is required by gqlgen
func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

// ID provides the UUID instead of the autoincrement ID.
func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// Organizations retrieves the list of Organizations to which the queried user is associated
func (r *userResolver) Organizations(ctx context.Context, obj *models.User) ([]models.Organization, error) {
	if obj == nil {
		return nil, nil
	}

	organizations, err := obj.GetOrganizations(models.Tx(ctx))
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetUserOrganizations")
	}

	return organizations, nil
}

// Requests retrieves the list of Requests associated with the queried user, where association is defined by the given `role`.
func (r *userResolver) Requests(ctx context.Context, obj *models.User, role RequestRole) ([]models.Request, error) {
	if obj == nil {
		return nil, nil
	}

	requests, err := obj.Requests(models.Tx(ctx), requestRoleMap[role])
	if err != nil {
		domain.NewExtra(ctx, "role", role)
		return nil, domain.ReportError(ctx, err, "GetUserRequests")
	}

	return requests, nil
}

// AvatarURL retrieves a URL for the user profile photo or avatar.
func (r *userResolver) AvatarURL(ctx context.Context, obj *models.User) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	photoURL, err := obj.GetPhotoURL(models.Tx(ctx))
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetUserPhotoURL")
	}

	return photoURL, nil
}

// PhotoID retrieves the ID for the user profile photo
func (r *userResolver) PhotoID(ctx context.Context, obj *models.User) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	if !obj.FileID.Valid {
		return nil, nil
	}

	photoID, err := obj.GetPhotoID(models.Tx(ctx))
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetUserPhotoID")
	}

	return photoID, nil
}

// Location retrieves the queried user's location.
func (r *userResolver) Location(ctx context.Context, obj *models.User) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	location, err := obj.GetLocation(models.Tx(ctx))
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetUserLocation")
	}

	return location, nil
}

// UnreadMessageCount calculates the number of unread messages for the queried user
func (r *userResolver) UnreadMessageCount(ctx context.Context, obj *models.User) (int, error) {
	if obj == nil {
		return 0, nil
	}
	mCounts, err := obj.UnreadMessageCount(models.Tx(ctx))
	if err != nil {
		return 0, domain.ReportError(ctx, err, "GetUserUnreadMessageCount")
	}
	total := 0
	for _, c := range mCounts {
		total += c.Count
	}

	return total, nil
}

// Preferences resolves the `preferences` property of the user query, retrieving the related records from the database
// and using them to hydrate a StandardPreferences struct.
func (r *userResolver) Preferences(ctx context.Context, obj *models.User) (*models.StandardPreferences, error) {
	if obj == nil {
		return nil, nil
	}

	standardPrefs, err := obj.GetPreferences(models.Tx(ctx))
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetUserPreferences")
	}

	return &standardPrefs, nil
}

// MeetingsAsParticipant returns all meetings in which the user is a participant
func (r *userResolver) MeetingsAsParticipant(ctx context.Context, obj *models.User) ([]models.Meeting, error) {
	if obj == nil {
		return nil, nil
	}

	meetings, err := obj.MeetingsAsParticipant(models.Tx(ctx))
	if err != nil {
		domain.NewExtra(ctx, "user", obj.UUID)
		return []models.Meeting{}, domain.ReportError(ctx, err, "User.MeetingsAsParticipant")
	}
	return meetings, nil
}

// Users retrieves a list of users
func (r *queryResolver) Users(ctx context.Context) ([]models.User, error) {
	currentUser := models.CurrentUser(ctx)

	role := currentUser.AdminRole
	if role != models.UserAdminRoleSuperAdmin {
		err := errors.New("insufficient permissions")
		domain.NewExtra(ctx, "role", role)
		return nil, domain.ReportError(ctx, err, "GetUsers.Unauthorized")
	}

	users := models.Users{}
	if err := users.All(models.Tx(ctx)); err != nil {
		return nil, domain.ReportError(ctx, err, "GetUsers")
	}

	return users, nil
}

// User retrieves a single user
func (r *queryResolver) User(ctx context.Context, id *string) (*models.User, error) {
	currentUser := models.CurrentUser(ctx)

	if id == nil {
		return &currentUser, nil
	}

	role := currentUser.AdminRole
	if role != models.UserAdminRoleSuperAdmin && currentUser.UUID.String() != *id {
		err := errors.New("insufficient permissions")
		domain.NewExtra(ctx, "role", role)
		return nil, domain.ReportError(ctx, err, "GetUser.Unauthorized")
	}

	dbUser := models.User{}
	if err := dbUser.FindByUUID(models.Tx(ctx), *id); err != nil {
		return &models.User{}, domain.ReportError(ctx, err, "GetUser")
	}

	return &dbUser, nil
}

// UpdateUser takes data from the GraphQL `UpdateUser` mutation and updates the database. If the
// user ID is provided and the current user is allowed to edit profiles, that user will be updated.
// Otherwise, the current authenticated user is updated.
func (r *mutationResolver) UpdateUser(ctx context.Context, input UpdateUserInput) (*models.User, error) {
	cUser := models.CurrentUser(ctx)
	var user models.User

	tx := models.Tx(ctx)
	if input.ID != nil {
		if err := user.FindByUUID(tx, *(input.ID)); err != nil {
			return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.NotFound")
		}
	} else {
		user = cUser
	}

	if cUser.AdminRole != models.UserAdminRoleSuperAdmin && cUser.ID != user.ID {
		err := errors.New("insufficient permissions")
		return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.Unauthorized")
	}

	if input.Nickname != nil {
		user.Nickname = *input.Nickname
	}

	var err error
	if input.PhotoID == nil {
		err = user.RemoveFile(tx)
	} else {
		_, err = user.AttachPhoto(tx, *input.PhotoID)
	}
	if err != nil {
		return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.UpdatePhoto")
	}

	if input.Location == nil {
		err = user.RemoveLocation(tx)
	} else {
		err = user.SetLocation(tx, convertLocation(*input.Location))
	}
	if err != nil {
		return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.SetLocationError")
	}

	if input.Preferences != nil {
		standardPrefs, err := convertUserPreferencesToStandardPreferences(input.Preferences)
		if err != nil {
			return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.PreferencesInput")
		}

		if _, err = user.UpdateStandardPreferences(tx, standardPrefs); err != nil {
			return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.Preferences")
		}
	} else {
		if err := user.RemovePreferences(tx); err != nil {
			return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.RemovePreferences")
		}
	}

	if err = user.Save(tx); err != nil {
		if strings.Contains(err.Error(), "Nickname must have a visible character") {
			return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.InvisibleNickname")
		}
		if strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_nickname_idx"`) {
			return &models.User{}, domain.ReportError(ctx, err, "UpdateUser.DuplicateNickname")
		}
		return &models.User{}, domain.ReportError(ctx, err, "UpdateUser")
	}

	return &user, nil
}

// getPublicProfiles converts a list of models.User to PublicProfile, hiding private profile information
func getPublicProfiles(ctx context.Context, users []models.User) []PublicProfile {
	profiles := make([]PublicProfile, len(users))
	for i, p := range users {
		user := p
		prof := getPublicProfile(ctx, &user)
		profiles[i] = *prof
	}
	return profiles
}

// getPublicProfile converts a models.User to a PublicProfile, which hides private profile information
func getPublicProfile(ctx context.Context, user *models.User) *PublicProfile {
	if user == nil {
		return &PublicProfile{}
	}

	url, err := user.GetPhotoURL(models.Tx(ctx))
	if err != nil {
		domain.NewExtra(ctx, "user", user.UUID)
		_ = domain.ReportError(ctx, err, "")
		return &PublicProfile{
			ID:       user.UUID.String(),
			Nickname: user.Nickname,
		}
	}

	return &PublicProfile{
		ID:        user.UUID.String(),
		Nickname:  user.Nickname,
		AvatarURL: url,
	}
}

func (r *Resolver) UserPreferences() UserPreferencesResolver {
	return &userPreferencesResolver{r}
}

type userPreferencesResolver struct{ *Resolver }

func (u userPreferencesResolver) Language(ctx context.Context, obj *models.StandardPreferences) (*PreferredLanguage, error) {
	language := PreferredLanguage(strings.ToUpper(obj.Language))
	return &language, nil
}

func (u userPreferencesResolver) WeightUnit(ctx context.Context, obj *models.StandardPreferences) (*PreferredWeightUnit, error) {
	unit := PreferredWeightUnit(strings.ToUpper(obj.WeightUnit))
	return &unit, nil
}
