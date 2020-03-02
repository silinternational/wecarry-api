package models

import (
	"database/sql"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestMeetingInvite_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		inv      MeetingInvite
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			inv: MeetingInvite{
				MeetingID: 1,
				InviterID: 1,
				Secret:    domain.GetUUID(),
				Email:     "foo@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing MeetingID",
			inv: MeetingInvite{
				InviterID: 1,
				Secret:    domain.GetUUID(),
				Email:     "foo@example.com",
			},
			wantErr:  true,
			errField: "meeting_id",
		},
		{
			name: "missing InviterID",
			inv: MeetingInvite{
				MeetingID: 1,
				Secret:    domain.GetUUID(),
				Email:     "foo@example.com",
			},
			wantErr:  true,
			errField: "inviter_id",
		},
		{
			name: "missing Secret",
			inv: MeetingInvite{
				MeetingID: 1,
				InviterID: 1,
				Email:     "foo@example.com",
			},
			wantErr:  true,
			errField: "secret",
		},
		{
			name: "missing Email",
			inv: MeetingInvite{
				MeetingID: 1,
				InviterID: 1,
				Secret:    domain.GetUUID(),
			},
			wantErr:  true,
			errField: "email",
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

func (ms *ModelSuite) TestMeetingInvite_Create() {
	meetings := createMeetingFixtures(ms.DB, 2).Meetings
	inviter := createUserFixtures(ms.DB, 1).Users[0]
	invite := MeetingInvite{
		MeetingID: meetings[0].ID,
		InviterID: inviter.ID,
		Email:     "existing@example.com",
	}
	ms.NoError(invite.Create())

	tests := []struct {
		name    string
		invite  MeetingInvite
		wantErr string
	}{
		{
			name: "good",
			invite: MeetingInvite{
				MeetingID: meetings[0].ID,
				InviterID: inviter.ID,
				Email:     "foo@example.com",
			},
		},
		{
			name:    "fail validation",
			invite:  MeetingInvite{},
			wantErr: "Email does not match the email format.",
		},
		{
			name: "bad foreign key",
			invite: MeetingInvite{
				MeetingID: 99999,
				InviterID: inviter.ID,
				Email:     "foo@example.com",
			},
			wantErr: "foreign key constraint",
		},
		{
			name: "don't fail on duplicate",
			invite: MeetingInvite{
				MeetingID: meetings[0].ID,
				InviterID: inviter.ID,
				Email:     "existing@example.com",
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			err := tt.invite.Create()
			if tt.wantErr != "" {
				ms.Error(err, `didn't get expected error: "%s"`, tt.wantErr)
				ms.Contains(err.Error(), tt.wantErr, "wrong error message")
				return
			}
			ms.NoError(err)
			ms.False(tt.invite.Secret == uuid.Nil, "didn't get a valid secret code")
		})
	}
}

func (ms *ModelSuite) TestMeetingInvite_Meeting() {
	meetings := createMeetingFixtures(ms.DB, 2).Meetings

	tests := []struct {
		name    string
		invite  MeetingInvite
		want    Meeting
		wantErr string
	}{
		{
			name: "good",
			invite: MeetingInvite{
				MeetingID: meetings[0].ID,
			},
			want: meetings[0],
		},
		{
			name: "bad",
			invite: MeetingInvite{
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

func (ms *ModelSuite) TestMeetingInvite_Inviter() {
	user := createUserFixtures(ms.DB, 2).Users[0]

	tests := []struct {
		name    string
		invite  MeetingInvite
		want    User
		wantErr string
	}{
		{
			name: "good",
			invite: MeetingInvite{
				InviterID: user.ID,
			},
			want: user,
		},
		{
			name:    "bad",
			invite:  MeetingInvite{},
			wantErr: "sql: no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.invite.Inviter()
			if tt.wantErr != "" {
				ms.Error(err, `didn't get expected error: "%s"`, tt.wantErr)
				ms.Contains(err.Error(), tt.wantErr, "wrong error message")
				return
			}
			ms.Equal(tt.want.UUID, got.UUID, "wrong user returned")
		})
	}
}

func (ms *ModelSuite) TestMeetingInvite_FindBySecret() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		meeting int
		email   string
		secret  string
		want    MeetingInvite
		wantErr string
	}{
		{
			name:    "bad - wrong meeting",
			meeting: f.Meetings[1].ID,
			email:   f.Users[2].Email,
			secret:  f.MeetingInvites[0].Secret.String(),
			wantErr: sql.ErrNoRows.Error(),
		},
		{
			name:    "bad - wrong email",
			meeting: f.Meetings[0].ID,
			email:   f.Users[1].Email,
			secret:  f.MeetingInvites[0].Secret.String(),
			wantErr: sql.ErrNoRows.Error(),
		},
		{
			name:    "bad - wrong secret",
			meeting: f.Meetings[0].ID,
			email:   f.Users[2].Email,
			secret:  f.MeetingInvites[1].Secret.String(),
			wantErr: sql.ErrNoRows.Error(),
		},
		{
			name:    "bad - meetingID=0",
			meeting: 0,
			email:   f.Users[2].Email,
			secret:  f.MeetingInvites[0].Secret.String(),
			wantErr: "invalid meeting ID",
		},
		{
			name:    "bad - empty email",
			meeting: f.Meetings[0].ID,
			email:   "",
			secret:  f.MeetingInvites[1].Secret.String(),
			wantErr: "empty email",
		},
		{
			name:    "bad - empty secret",
			meeting: f.Meetings[0].ID,
			email:   f.Users[2].Email,
			secret:  "",
			wantErr: "empty secret",
		},
		{
			name:    "good",
			meeting: f.Meetings[0].ID,
			email:   f.Users[2].Email,
			secret:  f.MeetingInvites[0].Secret.String(),
			want:    f.MeetingInvites[0],
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			var inv MeetingInvite
			err := inv.FindBySecret(tt.meeting, tt.email, tt.secret)
			if tt.wantErr != "" {
				ms.Error(err, `didn't get expected error: "%s"`, tt.wantErr)
				ms.Contains(err.Error(), tt.wantErr, "wrong error message")
				return
			}
			ms.Equal(tt.want.ID, inv.ID, "IsSecretValid returned the wrong invite")
		})
	}
}
