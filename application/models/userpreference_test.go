package models

import (
	"database/sql"
	"reflect"
	"runtime"
	"strconv"
	"testing"

	"github.com/gobuffalo/validate/v3"
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
			err := u.FindByUUID(ms.DB, test.uuid)
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
			err := test.pref.Save(ms.DB)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}
			ms.NoError(err)

			var u UserPreference
			ms.NoError(u.FindByUUID(ms.DB, test.pref.UUID.String()))
			ms.Equal(test.pref.UserID, u.UserID)
			ms.Equal(test.pref.Key, u.Key)
			ms.Equal(test.pref.Value, u.Value)
		})
	}
}

func (ms *ModelSuite) TestStandardPreferences_hydrateValues() {
	values := map[string]string{
		domain.UserPreferenceKeyLanguage:   domain.UserPreferenceLanguageFrench,
		domain.UserPreferenceKeyTimeZone:   "America/New_York",
		domain.UserPreferenceKeyWeightUnit: domain.UserPreferenceWeightUnitKGs,
	}
	sps := StandardPreferences{}
	sps.hydrateValues(values)

	want := StandardPreferences{
		Language:   values[domain.UserPreferenceKeyLanguage],
		TimeZone:   values[domain.UserPreferenceKeyTimeZone],
		WeightUnit: values[domain.UserPreferenceKeyWeightUnit],
	}

	ms.Equal(want, sps)
}

func (ms *ModelSuite) TestUserPreference_getPreferencesFieldsAndValidators() {
	sps := StandardPreferences{
		Language:   domain.UserPreferenceLanguageFrench,
		TimeZone:   "America/New_York",
		WeightUnit: domain.UserPreferenceWeightUnitKGs,
	}
	fAndVs := getPreferencesFieldsAndValidators(sps)

	wantValues := [3]string{sps.Language, sps.TimeZone, sps.WeightUnit}
	gotValues := [3]string{
		fAndVs[domain.UserPreferenceKeyLanguage].fieldValue,
		fAndVs[domain.UserPreferenceKeyTimeZone].fieldValue,
		fAndVs[domain.UserPreferenceKeyWeightUnit].fieldValue,
	}
	ms.Equal(wantValues, gotValues, "incorrect field values")

	wantValrs := [3]string{
		runtime.FuncForPC(reflect.ValueOf(domain.IsLanguageAllowed).Pointer()).Name(),
		runtime.FuncForPC(reflect.ValueOf(domain.IsTimeZoneAllowed).Pointer()).Name(),
		runtime.FuncForPC(reflect.ValueOf(domain.IsWeightUnitAllowed).Pointer()).Name(),
	}

	gotValrs := [3]string{
		runtime.FuncForPC(reflect.ValueOf(fAndVs[domain.UserPreferenceKeyLanguage].validator).Pointer()).Name(),
		runtime.FuncForPC(reflect.ValueOf(fAndVs[domain.UserPreferenceKeyTimeZone].validator).Pointer()).Name(),
		runtime.FuncForPC(reflect.ValueOf(fAndVs[domain.UserPreferenceKeyWeightUnit].validator).Pointer()).Name(),
	}

	ms.Equal(wantValrs, gotValrs, "incorrect validators")
}

func (ms *ModelSuite) TestUserPreference_updateForUserByKey() {
	t := ms.T()

	f := CreateUserFixtures_TestGetPreference(ms)

	tests := []struct {
		name  string
		user  User
		key   string
		value string
		want  UserPreference
	}{
		{
			name:  "Change Lang to French",
			user:  f.Users[0],
			key:   domain.UserPreferenceKeyLanguage,
			value: domain.UserPreferenceLanguageFrench,
			want: UserPreference{
				ID:     f.UserPreferences[0].ID,
				UserID: f.Users[0].ID,
				Key:    domain.UserPreferenceKeyLanguage,
				Value:  domain.UserPreferenceLanguageFrench,
			},
		},
		{
			name:  "Leave KGs unchanged",
			user:  f.Users[0],
			key:   domain.UserPreferenceKeyWeightUnit,
			value: domain.UserPreferenceWeightUnitKGs,
			want: UserPreference{
				ID:     f.UserPreferences[1].ID,
				UserID: f.Users[0].ID,
				Key:    domain.UserPreferenceKeyWeightUnit,
				Value:  domain.UserPreferenceWeightUnitKGs,
			},
		},
		{
			name:  "Add French",
			user:  f.Users[1],
			key:   domain.UserPreferenceKeyLanguage,
			value: domain.UserPreferenceLanguageFrench,
			want: UserPreference{
				UserID: f.Users[1].ID,
				Key:    domain.UserPreferenceKeyLanguage,
				Value:  domain.UserPreferenceLanguageFrench,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := UserPreference{}
			err := p.updateForUserByKey(ms.DB, test.user, test.key, test.value)
			ms.NoError(err)

			if test.want.ID > 0 {
				ms.Equal(test.want.ID, p.ID, "incorrect ID result from updatePreferenceByKey()")
			} else {
				ms.Greater(p.ID, 0, "non-positive ID from updatePreferenceByKey()")
			}
			ms.Equal(test.want.Key, p.Key, "incorrect key result from updatePreferenceByKey()")
			ms.Equal(test.want.Value, p.Value, "incorrect value result for "+test.want.Key)
		})
	}
}

func (ms *ModelSuite) TestUserPreference_remove() {
	t := ms.T()

	user := createUserFixtures(ms.DB, 1).Users[0]
	userPreference := UserPreference{
		UserID: user.ID,
		Key:    domain.UserPreferenceKeyLanguage,
		Value:  domain.UserPreferenceLanguageEnglish,
	}
	mustCreate(ms.DB, &userPreference)

	tests := []struct {
		name string
		user User
		key  string
	}{
		{
			name: "Remove language",
			user: user,
			key:  domain.UserPreferenceKeyLanguage,
		},
		{
			name: "Remove nonexistent key",
			user: user,
			key:  domain.UserPreferenceKeyTimeZone,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := UserPreference{}
			err := p.remove(ms.DB, test.user, test.key)
			ms.NoError(err, "unexpected error from UserPreference.remove()")

			err = ms.DB.Where("user_id = ? AND key = ?", test.user.ID, test.key).First(&p)
			ms.Error(err, "expected to get an error while finding deleted key")
			ms.Equal(sql.ErrNoRows, err, "unexpected error type while finding deleted key")
		})
	}
}
