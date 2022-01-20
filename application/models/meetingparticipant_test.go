package models

import (
	"sort"
	"testing"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
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
	meetings := createMeetingFixtures(ms.DB, 2).Meetings

	tests := []struct {
		name        string
		participant MeetingParticipant
		want        Meeting
		wantErr     string
	}{
		{
			name: "good",
			participant: MeetingParticipant{
				MeetingID: meetings[0].ID,
			},
			want: meetings[0],
		},
		{
			name: "bad",
			participant: MeetingParticipant{
				MeetingID: 0,
			},
			wantErr: "sql: no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.participant.Meeting(ms.DB)
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
		name        string
		participant MeetingParticipant
		want        User
		wantErr     string
	}{
		{
			name: "good",
			participant: MeetingParticipant{
				UserID: user.ID,
			},
			want: user,
		},
		{
			name:        "bad",
			participant: MeetingParticipant{},
			wantErr:     "sql: no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.participant.User(ms.DB)
			if tt.wantErr != "" {
				ms.Error(err, `didn't get expected error: "%s"`, tt.wantErr)
				ms.Contains(err.Error(), tt.wantErr, "wrong error message")
				return
			}
			ms.Equal(tt.want.UUID, got.UUID, "wrong user returned")
		})
	}
}

func (ms *ModelSuite) TestMeetingParticipant_FindOrCreate() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		user    User
		meeting Meeting
		invite  *MeetingInvite
		code    nulls.UUID
		userIDs []int
		wantErr string
	}{
		{
			name:    "invalid invite code",
			user:    f.Users[4],
			meeting: f.Meetings[0],
			code:    nulls.NewUUID(domain.GetUUID()),
			wantErr: "CreateMeetingParticipant.InvalidSecret",
		},
		{
			name:    "creator self-join",
			user:    f.Users[0],
			meeting: f.Meetings[0],
			code:    nulls.UUID{},
			userIDs: []int{f.Users[0].ID, f.Users[1].ID, f.Users[2].ID, f.Users[3].ID},
		},
		{
			name:    "organizer re-join",
			user:    f.Users[1],
			meeting: f.Meetings[0],
			code:    nulls.UUID{},
			userIDs: []int{f.Users[0].ID, f.Users[1].ID, f.Users[2].ID, f.Users[3].ID},
		},
		{
			name:    "participant re-join",
			user:    f.Users[2],
			meeting: f.Meetings[0],
			code:    nulls.UUID{},
			userIDs: []int{f.Users[0].ID, f.Users[1].ID, f.Users[2].ID, f.Users[3].ID},
		},
		{
			name:    "new participant - invite code",
			user:    f.Users[4],
			meeting: f.Meetings[0],
			code:    nulls.NewUUID(f.MeetingInvites[1].Secret),
			invite:  &f.MeetingInvites[1],
			userIDs: []int{f.Users[0].ID, f.Users[1].ID, f.Users[2].ID, f.Users[3].ID, f.Users[4].ID},
		},
		{
			name:    "new participant - meeting code",
			user:    f.Users[5],
			meeting: f.Meetings[0],
			code:    nulls.NewUUID(f.Meetings[0].InviteCode.UUID),
			userIDs: []int{f.Users[0].ID, f.Users[1].ID, f.Users[2].ID, f.Users[3].ID, f.Users[4].ID, f.Users[5].ID},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			// setup
			var code *string
			if tt.code.Valid {
				c := tt.code.UUID.String()
				code = &c
			} else {
				code = nil
			}

			// execute
			var p MeetingParticipant
			appErr := p.FindOrCreate(ms.DB, tt.meeting, tt.user, code)

			// verify
			if tt.wantErr != "" {
				ms.NotNil(appErr, "did not get expected error")
				ms.Contains(appErr.Key, tt.wantErr)
				return
			}
			ms.Nil(appErr, "unexpected error")

			participants, err := tt.meeting.Participants(ms.DB, f.Users[0])
			ms.NoError(err)

			ids := make([]int, len(participants))
			for i := range participants {
				ids[i] = participants[i].UserID
			}
			sort.Ints(ids)

			ms.Equal(tt.userIDs, ids)

			if tt.invite == nil {
				return
			}

			var invite MeetingInvite
			err = ms.DB.Find(&invite, tt.invite.ID)
			ms.NoError(err, "error finding MeetingInvite to validate results")

			ms.Equal(tt.user.UUID, invite.UserID.UUID, "incorrect invite userID")

			// teardown
		})
	}
}
