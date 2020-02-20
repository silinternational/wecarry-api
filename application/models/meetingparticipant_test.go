package models

import (
	"testing"
)

func (ms *ModelSuite) TestMeetingParticipant_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		inv      MeetingParticipant
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			inv: MeetingParticipant{
				MeetingID: 1,
				UserID:    1,
			},
			wantErr: false,
		},
		{
			name: "missing MeetingID",
			inv: MeetingParticipant{
				UserID: 1,
			},
			wantErr:  true,
			errField: "meeting_id",
		},
		{
			name: "missing UserID",
			inv: MeetingParticipant{
				MeetingID: 1,
			},
			wantErr:  true,
			errField: "user_id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.inv.Validate(DB)
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

func (ms *ModelSuite) TestMeetingParticipant_Meeting() {
	meetings := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		invite  MeetingParticipant
		want    Meeting
		wantErr string
	}{
		{
			name: "good",
			invite: MeetingParticipant{
				MeetingID: meetings[0].ID,
			},
			want: meetings[0],
		},
		{
			name: "bad",
			invite: MeetingParticipant{
				MeetingID: 0,
			},
			wantErr: "sql: no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.invite.Meeting()
			if tt.wantErr != "" {
				ms.Error(err, `didn't get expected error: "%s"`, tt.wantErr)
				ms.Contains(err.Error(), tt.wantErr, "wrong error message")
				return
			}
			ms.Equal(tt.want.UUID, got.UUID, "wrong meeting returned")
		})
	}
}

func (ms *ModelSuite) TestMeetingParticipant_User() {
	user := createUserFixtures(ms.DB, 2).Users[0]

	tests := []struct {
		name    string
		invite  MeetingParticipant
		want    User
		wantErr string
	}{
		{
			name: "good",
			invite: MeetingParticipant{
				UserID: user.ID,
			},
			want: user,
		},
		{
			name:    "bad",
			invite:  MeetingParticipant{},
			wantErr: "sql: no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.invite.User()
			if tt.wantErr != "" {
				ms.Error(err, `didn't get expected error: "%s"`, tt.wantErr)
				ms.Contains(err.Error(), tt.wantErr, "wrong error message")
				return
			}
			ms.Equal(tt.want.UUID, got.UUID, "wrong user returned")
		})
	}
}
