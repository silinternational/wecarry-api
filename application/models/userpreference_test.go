package models

import (
	"strconv"
	"testing"

	"github.com/gobuffalo/validate"
	"github.com/silinternational/wecarry-api/domain"
)

type UserPreferenceFixtures struct {
	Users
	UserPreferences
}

func (ms *ModelSuite) TestUserPreference_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		pref     UserPreference
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			pref: UserPreference{
				UUID:   domain.GetUUID(),
				UserID: 1,
				Key:    "key",
				Value:  "value",
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			pref: UserPreference{
				UserID: 1,
				Key:    "key",
				Value:  "value",
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "missing user_id",
			pref: UserPreference{
				UUID:  domain.GetUUID(),
				Key:   "key",
				Value: "value",
			},
			wantErr:  true,
			errField: "user_id",
		},
		{
			name: "missing key",
			pref: UserPreference{
				UUID:   domain.GetUUID(),
				UserID: 1,
				Value:  "value",
			},
			wantErr:  true,
			errField: "key",
		},
		{
			name: "missing value",
			pref: UserPreference{
				UUID:   domain.GetUUID(),
				UserID: 1,
				Key:    "key",
			},
			wantErr:  true,
			errField: "value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.pref.Validate(DB)
			if test.wantErr {
				ms.True(vErr.Count() != 0, "Expected an error, but did not get one")
				ms.True(len(vErr.Get(test.errField)) > 0,
					"Expected an error on field %v, but got none (errors: %v)",
					test.errField, vErr.Errors)
				return
			}
			ms.False(vErr.HasAny(), "Unexpected error: %v", vErr)
		})
	}
}

func createFixturesForUserPreferenceFindByUUID(ms *ModelSuite) UserPreferenceFixtures {
	unique := domain.GetUUID().String()
	user := User{UUID: domain.GetUUID(), Email: unique + "_user@example.com", Nickname: unique + "_User"}
	createFixture(ms, &user)

	userPreferences := make(UserPreferences, 2)
	for i := range userPreferences {
		userPreferences[i] = UserPreference{
			UUID:   domain.GetUUID(),
			UserID: user.ID,
			Key:    "k",
			Value:  "v",
		}
		createFixture(ms, &userPreferences[i])
	}

	return UserPreferenceFixtures{
		Users:           Users{user},
		UserPreferences: userPreferences,
	}
}

func (ms *ModelSuite) TestUserPreference_FindByUUID() {
	t := ms.T()
	f := createFixturesForUserPreferenceFindByUUID(ms)
	tests := []struct {
		name    string
		uuid    string
		wantErr string
	}{
		{name: "good", uuid: f.UserPreferences[0].UUID.String()},
		{name: "bad", wantErr: "user preference uuid must not be blank"},
		{name: "not found", uuid: domain.GetUUID().String(), wantErr: "sql: no rows in result set"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var u UserPreference
			err := u.FindByUUID(test.uuid)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}
			ms.NoError(err)
			ms.Equal(test.uuid, u.UUID.String())
		})
	}
}

func createFixturesForUserPreferenceSave(ms *ModelSuite) UserPreferenceFixtures {
	unique := domain.GetUUID().String()
	user := User{UUID: domain.GetUUID(), Email: unique + "_user@example.com", Nickname: unique + "_User"}
	createFixture(ms, &user)

	userPreferences := make(UserPreferences, 2)
	for i := range userPreferences {
		userPreferences[i] = UserPreference{
			UserID: user.ID,
			Key:    "key" + strconv.Itoa(i),
			Value:  "v",
		}
	}
	createFixture(ms, &userPreferences[0])

	return UserPreferenceFixtures{
		Users:           Users{user},
		UserPreferences: userPreferences,
	}
}

func (ms *ModelSuite) TestUserPreference_Save() {
	t := ms.T()
	f := createFixturesForUserPreferenceSave(ms)
	tests := []struct {
		name    string
		pref    UserPreference
		wantErr string
	}{
		{name: "update", pref: f.UserPreferences[0]},
		{name: "create", pref: f.UserPreferences[1]},
		{name: "bad", wantErr: "can not be blank"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.pref.Save()
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}
			ms.NoError(err)

			var u UserPreference
			ms.NoError(u.FindByUUID(test.pref.UUID.String()))
			ms.Equal(test.pref.UserID, u.UserID)
			ms.Equal(test.pref.Key, u.Key)
			ms.Equal(test.pref.Value, u.Value)
		})
	}
}
