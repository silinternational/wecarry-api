package models

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"

	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestWatch_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		watch    Watch
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			watch: Watch{
				UUID:    domain.GetUUID(),
				OwnerID: 1,
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			watch: Watch{
				OwnerID: 1,
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "missing owner_id",
			watch: Watch{
				UUID: domain.GetUUID(),
			},
			wantErr:  true,
			errField: "owner_id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.watch.Validate(DB)
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

// createWatchFixtures prepares Watch fixtures that will not match anything until modified to meet the needs of a test
func createWatchFixtures(tx *pop.Connection, users Users) Watches {
	watches := make(Watches, len(users)*2)
	for i := range watches {
		watches[i].UUID = domain.GetUUID()
		watches[i].OwnerID = users[i/2].ID
		watches[i].Name = domain.GetUUID().String()
		watches[i].SearchText = nulls.NewString(watches[i].Name)
		mustCreate(tx, &watches[i])
	}
	return watches
}

func (ms *ModelSuite) TestWatch_FindByUUID() {
	t := ms.T()

	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 2).Users)

	tests := []struct {
		name    string
		uuid    string
		want    Watch
		wantErr string
	}{
		{name: "user 0", uuid: watches[0].UUID.String(), want: watches[0]},
		{name: "user 1", uuid: watches[2].UUID.String(), want: watches[2]},
		{name: "blank uuid", uuid: "", wantErr: "watch uuid must not be blank"},
		{name: "wrong uuid", uuid: domain.GetUUID().String(), wantErr: "no rows in result set"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var watch Watch
			err := watch.FindByUUID(test.uuid)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "wrong error type")
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(test.want.UUID, watch.UUID, "incorrect uuid")
		})
	}
}

func (ms *ModelSuite) TestWatches_FindByUser() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	watches := createWatchFixtures(ms.DB, users)
	noWatches := createUserFixtures(ms.DB, 1).Users[0]

	tests := []struct {
		name string
		user User
		want Watches
	}{
		{name: "user 0", user: users[0], want: Watches{watches[1], watches[0]}},
		{name: "user 1", user: users[1], want: Watches{watches[3], watches[2]}},
		{name: "no watches", user: noWatches, want: Watches{}},
		{name: "wrong user", user: User{}, want: Watches{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Watches{}
			err := got.FindByUser(test.user)
			ms.NoError(err, "unexpected error")

			gotIDs := make([]string, len(got))
			for i := range got {
				gotIDs[i] = got[i].UUID.String()
			}

			wantIDs := make([]string, len(test.want))
			for i := range test.want {
				wantIDs[i] = test.want[i].UUID.String()
			}

			ms.Equal(wantIDs, gotIDs, "wrong list of watches")
		})
	}
}

func (ms *ModelSuite) TestWatch_GetOwner() {
	users := createUserFixtures(ms.DB, 2).Users
	watches := createWatchFixtures(ms.DB, users)

	owner, err := watches[0].GetOwner()
	ms.NoError(err, "unexpected error")
	ms.Equal(users[0].UUID, owner.UUID, "incorrect owner")
}

func (ms *ModelSuite) TestWatch_GetSetLocation() {
	newLoc := createLocationFixtures(ms.DB, 1)[0]
	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 1).Users)

	err := watches[0].SetDestination(newLoc)
	ms.NoError(err, "unexpected error from SetDestination()")

	got, err := watches[0].GetDestination()
	ms.NoError(err, "unexpected error from GetLocation()")
	ms.Equal(newLoc.Country, got.Country, "country doesn't match")
	ms.Equal(newLoc.Description, got.Description, "description doesn't match")
	ms.InDelta(newLoc.Latitude.Float64, got.Latitude.Float64, 0.0001, "latitude doesn't match")
	ms.InDelta(newLoc.Longitude.Float64, got.Longitude.Float64, 0.0001, "longitude doesn't match")
}

func (ms *ModelSuite) TestWatch_Meeting() {
	users := createUserFixtures(ms.DB, 2).Users
	watches := createWatchFixtures(ms.DB, users)
	meeting := createMeetingFixtures(ms.DB, 1).Meetings[0]
	watches[1].MeetingID = nulls.NewInt(meeting.ID)
	ms.NoError(watches[1].Update())

	tests := []struct {
		name     string
		testUser User
		watch    Watch
		want     *Meeting
	}{
		{
			name:     "no meeting",
			testUser: users[0],
			watch:    watches[0],
			want:     nil,
		},
		{
			name:     "has meeting",
			testUser: users[0],
			watch:    watches[1],
			want:     &meeting,
		},
		{
			name:     "not authorized",
			testUser: users[1],
			watch:    watches[1],
			want:     nil,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.watch.Meeting(createTestContext(tt.testUser))
			ms.NoError(err)
			if tt.want == nil {
				ms.Nil(got)
				return
			}
			ms.NotNil(got, "Watch.Meeting() returned nil")
			ms.Equal(tt.want.ID, got.ID)
		})
	}
}

func (ms *ModelSuite) TestWatch_requestMatches() {
	requests := createRequestFixtures(ms.DB, 1, false)
	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 2).Users)

	// watch 0 matches on text, but doesn't match size
	tiny := RequestSizeTiny
	watches[0].Size = &tiny
	requestTitle := requests[0].Title
	watches[0].SearchText = nulls.NewString(requestTitle[:len(requestTitle)-1])
	ms.NoError(watches[1].Update())

	// watch 1 matches on text and size
	small := RequestSizeSmall
	watches[1].Size = &small
	watches[1].SearchText = nulls.NewString(requestTitle[:len(requestTitle)-1])
	ms.NoError(watches[1].Update())

	// watch 2 matches on neither text nor size
	watches[2].Size = &tiny
	watches[2].SearchText = nulls.NewString("not going to match this")
	ms.NoError(watches[2].Update())

	tests := []struct {
		name    string
		watch   *Watch
		request Request
		want    bool
	}{
		{
			name:  "nil",
			watch: nil,
			want:  false,
		},
		{
			name:    "one matching field, one mismatching field",
			watch:   &watches[0],
			request: requests[0],
			want:    false,
		},
		{
			name:    "two matching fields",
			watch:   &watches[1],
			request: requests[0],
			want:    true,
		},
		{
			name:    "two mis-matching fields",
			watch:   &watches[2],
			request: requests[0],
			want:    false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			if got := tt.watch.matchesRequest(tt.request); got != tt.want {
				t.Errorf("matchesRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func (ms *ModelSuite) TestWatch_destinationMatches() {
	requests := createRequestFixtures(ms.DB, 1, false)
	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 2).Users)

	dest, err := requests[0].GetDestination()
	ms.NoError(err)
	ms.NoError(dest.Create())
	ms.NoError(watches[0].SetDestination(*dest))

	ms.NoError(watches[1].SetDestination(Location{Country: "XX", Description: "-"}))

	tests := []struct {
		name    string
		watch   *Watch
		request Request
		want    bool
	}{
		{
			name:    "match",
			watch:   &watches[0],
			request: requests[0],
			want:    true,
		},
		{
			name:    "not match",
			watch:   &watches[1],
			request: requests[0],
			want:    false,
		},
		{
			name:    "destination is nil",
			watch:   &watches[2],
			request: requests[0],
			want:    true,
		},
		{
			name:    "watch is nil",
			watch:   nil,
			request: requests[0],
			want:    false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			ms.Equal(tt.want, tt.watch.destinationMatches(tt.request))
		})
	}
}

func (ms *ModelSuite) TestWatch_originMatches() {
	requests := createRequestFixtures(ms.DB, 1, false)
	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 2).Users)

	origin, err := requests[0].GetOrigin()
	ms.NoError(err)
	ms.NoError(origin.Create())
	ms.NoError(watches[0].SetOrigin(*origin))

	ms.NoError(watches[1].SetOrigin(Location{Country: "XX", Description: "-"}))

	tests := []struct {
		name    string
		watch   *Watch
		request Request
		want    bool
	}{
		{
			name:    "match",
			watch:   &watches[0],
			request: requests[0],
			want:    true,
		},
		{
			name:    "not match",
			watch:   &watches[1],
			request: requests[0],
			want:    false,
		},
		{
			name:    "origin is nil",
			watch:   &watches[2],
			request: requests[0],
			want:    true,
		},
		{
			name:    "watch is nil",
			watch:   nil,
			request: requests[0],
			want:    false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			ms.Equal(tt.want, tt.watch.originMatches(tt.request))
		})
	}
}

func (ms *ModelSuite) TestWatch_sizeMatches() {
	requests := createRequestFixtures(ms.DB, 1, false)
	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 2).Users)

	// don't need to save these changes because sizeMatches doesn't access the database
	requestSize := requests[0].Size // RequestSizeSmall
	watches[0].Size = &requestSize
	tiny := RequestSizeTiny
	watches[1].Size = &tiny

	tests := []struct {
		name    string
		watch   *Watch
		request Request
		want    bool
	}{
		{
			name:    "match",
			watch:   &watches[0],
			request: requests[0],
			want:    true,
		},
		{
			name:    "not match",
			watch:   &watches[1],
			request: requests[0],
			want:    false,
		},
		{
			name:    "size is nil",
			watch:   &watches[2],
			request: requests[0],
			want:    true,
		},
		{
			name:    "watch is nil",
			watch:   nil,
			request: requests[0],
			want:    false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			ms.Equal(tt.want, tt.watch.sizeMatches(tt.request))
		})
	}
}

func (ms *ModelSuite) TestWatch_meetingMatches() {
	requests := createRequestFixtures(ms.DB, 1, false)
	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 2).Users)

	// don't need to save these changes because meetingMatches doesn't access the database
	watches[0].MeetingID = nulls.NewInt(1)
	watches[1].MeetingID = nulls.NewInt(2)
	requests[0].MeetingID = nulls.NewInt(1)

	tests := []struct {
		name    string
		watch   *Watch
		request Request
		want    bool
	}{
		{
			name:    "match",
			watch:   &watches[0],
			request: requests[0],
			want:    true,
		},
		{
			name:    "not match",
			watch:   &watches[1],
			request: requests[0],
			want:    false,
		},
		{
			name:    "meeting is nil",
			watch:   &watches[2],
			request: requests[0],
			want:    true,
		},
		{
			name:    "watch is nil",
			watch:   nil,
			request: requests[0],
			want:    false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			ms.Equal(tt.want, tt.watch.meetingMatches(tt.request))
		})
	}
}

func (ms *ModelSuite) TestWatch_textMatches() {
	requests := createRequestFixtures(ms.DB, 1, false)
	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 3).Users)

	requestDescription := requests[0].Description.String
	watches[0].SearchText = nulls.NewString(requestDescription[:len(requestDescription)-1])
	ms.NoError(watches[0].Update())

	requestTitle := requests[0].Title
	watches[1].SearchText = nulls.NewString(requestTitle[:len(requestTitle)-1])
	ms.NoError(watches[1].Update())

	requestCreator, err := requests[0].Creator()
	ms.NoError(err)
	requestCreatorNickname := requestCreator.Nickname
	watches[2].SearchText = nulls.NewString(requestCreatorNickname[:5])
	ms.NoError(watches[2].Update())

	watches[3].SearchText = nulls.NewString("not a match for anything")
	ms.NoError(watches[3].Update())

	watches[4].SearchText = nulls.String{}
	ms.NoError(watches[4].Update())

	tests := []struct {
		name    string
		watch   *Watch
		request Request
		want    bool
	}{
		{
			name:    "match description",
			watch:   &watches[0],
			request: requests[0],
			want:    true,
		},
		{
			name:    "match title",
			watch:   &watches[1],
			request: requests[0],
			want:    true,
		},
		{
			name:    "match nickname",
			watch:   &watches[2],
			request: requests[0],
			want:    true,
		},
		{
			name:    "not match",
			watch:   &watches[3],
			request: requests[0],
			want:    false,
		},
		{
			name:    "search is nil",
			watch:   &watches[4],
			request: requests[0],
			want:    true,
		},
		{
			name:    "watch is nil",
			watch:   nil,
			request: requests[0],
			want:    false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			ms.Equal(tt.want, tt.watch.textMatches(tt.request))
		})
	}
}
