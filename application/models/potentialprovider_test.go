package models

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
)

func (ms *ModelSuite) TestPotentialProviders_FindUsersByRequestID() {
	f := createPotentialProvidersFixtures(ms)
	users := f.Users
	users[3].AdminRole = UserAdminRoleSuperAdmin
	ms.NoError(ms.DB.Save(&users[3]), "error making user fixture a SuperAdmin")

	requests := f.Requests
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name    string
		request Request
		user    User
		wantIDs []int
	}{
		{
			name:    "requester as current user",
			request: requests[0],
			user:    users[0],
			wantIDs: []int{pps[0].UserID, pps[1].UserID, pps[2].UserID},
		},
		{
			name:    "potential provider as current user",
			request: requests[0],
			user:    users[1],
			wantIDs: []int{pps[0].UserID},
		},
		{
			name:    "current user is SuperAdmin",
			request: requests[0],
			user:    users[3],
			wantIDs: []int{pps[0].UserID, pps[1].UserID, pps[2].UserID},
		},
		{
			name:    "non potential provider as current user",
			request: requests[1],
			user:    users[1],
			wantIDs: []int{},
		},
		{
			name:    "empty current user",
			request: requests[1],
			user:    User{},
			wantIDs: []int{pps[3].UserID, pps[4].UserID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers := PotentialProviders{}
			users, err := providers.FindUsersByRequestID(ms.DB, tt.request, tt.user)
			ms.NoError(err, "unexpected error")
			ids := make([]int, len(users))
			for i, u := range users {
				ids[i] = u.ID
			}

			ms.Equal(tt.wantIDs, ids)
		})
	}
}

func (ms *ModelSuite) TestPotentialProvider_FindWithRequestUUIDAndUserUUID() {
	f := createPotentialProvidersFixtures(ms)
	requests := f.Requests
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		request     Request
		ppUserUUID  uuid.UUID
		currentUser User
		wantID      int
		wantErr     string
	}{
		{
			name:        "request Creator as current user",
			request:     requests[0],
			ppUserUUID:  users[1].UUID,
			currentUser: users[1],
			wantID:      pps[0].ID,
		},
		{
			name:        "potential provider as current user",
			request:     requests[0],
			ppUserUUID:  users[2].UUID,
			currentUser: users[2],
			wantID:      pps[1].ID,
		},
		{
			name:        "current user is not potential provider",
			request:     requests[1],
			ppUserUUID:  users[3].UUID,
			currentUser: users[2],
			wantID:      pps[4].ID,
			wantErr:     "user not allowed to access PotentialProvider",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := PotentialProvider{}
			err := provider.FindWithRequestUUIDAndUserUUID(ms.DB, tt.request.UUID.String(),
				tt.ppUserUUID.String(), tt.currentUser)

			if tt.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tt.wantErr, "didn't get the expected error")
				return
			}

			ms.NoError(err, "unexpected error")
			ms.Equal(tt.wantID, provider.ID)
		})
	}
}

func (ms *ModelSuite) TestPotentialProviders_DestroyAllWithRequestUUID() {
	f := createPotentialProvidersFixtures(ms)
	requests := f.Requests
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		currentUser User
		request     Request
		wantIDs     []int
		wantErr     string
	}{
		{
			name:        "good: Request Creator as current user",
			currentUser: users[0],
			request:     requests[0],
			wantIDs:     []int{pps[3].ID, pps[4].ID},
		},
		{
			name:        "bad: current user is potential provider but not Request Creator",
			currentUser: users[2],
			request:     requests[0],
			wantErr: fmt.Sprintf(`user %v has insufficient permissions to destroy PotentialProviders for Request %v`,
				users[2].ID, requests[0].ID),
		},
		{
			name:        "bad: current user is not Request Creator",
			currentUser: users[1],
			request:     requests[1],
			wantErr: fmt.Sprintf(`user %v has insufficient permissions to destroy PotentialProviders for Request %v`,
				users[1].ID, requests[1].ID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			providers := PotentialProviders{}
			err := providers.DestroyAllWithRequestUUID(ms.DB, test.request.UUID.String(), test.currentUser)

			if test.wantErr != "" {
				ms.Error(err, "did not get error as expected")
				ms.Equal(test.wantErr, err.Error(), "wrong error message")
				return
			}

			ms.NoError(err, "unexpected error")

			var provs PotentialProviders
			err = DB.All(&provs)
			ms.NoError(err, "error just getting PotentialProviders back out of the DB.")

			pIDs := make([]int, len(provs))
			for i, p := range provs {
				pIDs[i] = p.ID
			}

			ms.Equal(test.wantIDs, pIDs)
		})
	}
}

func (ms *ModelSuite) TestNewWithRequestUUID() {
	f := createPotentialProvidersFixtures(ms)
	users := f.Users
	requests := f.Requests

	t := ms.T()
	tests := []struct {
		name    string
		request Request
		userID  int
		wantIDs []int
		wantErr string
	}{
		{
			name:    "bad - using request's CreatedBy",
			request: requests[0],
			userID:  users[0].ID,
			wantErr: "the PotentialProvider User must not be the Request's Receiver",
		},
		{
			name:    "good - second request second user",
			request: requests[1],
			userID:  users[1].ID,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{}
			err := provider.NewWithRequestUUID(ms.DB, test.request.UUID.String(), test.userID)
			if test.wantErr != "" {
				ms.Error(err, "expected an error but did not get one")
				ms.Equal(test.wantErr, err.Error(), "incorrect error message")
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(test.request.ID, provider.RequestID, "incorrect Request ID")
			ms.Equal(test.userID, provider.UserID, "incorrect User ID")
		})
	}
}

func (ms *ModelSuite) TestPotentialProvider_Validate() {
	f := createPotentialProvidersFixtures(ms)
	users := f.Users
	requests := f.Requests

	t := ms.T()
	tests := []struct {
		name      string
		requestID int
		userID    int
		wantIDs   []int
		wantErrs  map[string][]string
	}{
		{
			name:      "good - second request second user",
			requestID: requests[1].ID,
			userID:    users[1].ID,
			wantErrs:  map[string][]string{},
		},
		{
			name:      "bad - duplicate",
			requestID: requests[1].ID,
			userID:    users[3].ID,
			wantErrs: map[string][]string{
				"unique_together": {
					fmt.Sprintf("Duplicate potential provider exists with RequestID: %v and UserID: %v", requests[1].ID, users[3].ID),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{RequestID: test.requestID, UserID: test.userID}
			vErrors, err := provider.Validate(ms.DB)

			ms.NoError(err, "unexpected error")
			ms.Equal(test.wantErrs, vErrors.Errors, "incorrect validation errors")
		})
	}
}
