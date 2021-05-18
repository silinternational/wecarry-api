package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
)

type meetingFixtures struct {
	Meetings
	MeetingInvites
	MeetingParticipants
	Users
}

// TestMeeting_Validate ensures errors are thrown for missing required fields
func (ms *ModelSuite) TestMeeting_Validate() {
	t := ms.T()
	now := time.Now()

	tests := []struct {
		name     string
		meeting  Meeting
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			meeting: Meeting{
				UUID:        domain.GetUUID(),
				Name:        "A Meeting",
				CreatedByID: 1,
				LocationID:  1,
				StartDate:   now,
				EndDate:     now,
			},
			wantErr: false,
		},
		{
			name: "missing created_by",
			meeting: Meeting{
				UUID:       domain.GetUUID(),
				Name:       "A Meeting",
				LocationID: 1,
				StartDate:  now,
				EndDate:    now,
			},
			wantErr:  true,
			errField: "created_by_id",
		},
		{
			name: "missing location",
			meeting: Meeting{
				UUID:        domain.GetUUID(),
				Name:        "A Meeting",
				CreatedByID: 1,
				StartDate:   now,
				EndDate:     now,
			},
			wantErr:  true,
			errField: "location_id",
		},
		{
			name: "missing start_date",
			meeting: Meeting{
				UUID:        domain.GetUUID(),
				Name:        "A Meeting",
				CreatedByID: 1,
				LocationID:  1,
				EndDate:     now,
			},
			wantErr:  true,
			errField: "start_date",
		},
		{
			name: "missing end_date",
			meeting: Meeting{
				UUID:        domain.GetUUID(),
				Name:        "A Meeting",
				CreatedByID: 1,
				LocationID:  1,
				StartDate:   now,
			},
			wantErr:  true,
			errField: "end_date",
		},
		{
			name: "missing uuid",
			meeting: Meeting{
				Name:        "A Meeting",
				CreatedByID: 1,
				LocationID:  1,
				StartDate:   now,
				EndDate:     now,
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "dates the same",
			meeting: Meeting{
				Name:        "A Meeting",
				UUID:        domain.GetUUID(),
				CreatedByID: 1,
				LocationID:  1,
				StartDate:   now,
				EndDate:     now,
			},
			wantErr: false,
		},
		{
			name: "dates out of order",
			meeting: Meeting{
				Name:        "A Meeting",
				UUID:        domain.GetUUID(),
				CreatedByID: 1,
				LocationID:  1,
				StartDate:   time.Now().Add(time.Duration(domain.DurationDay)),
				EndDate:     now,
			},
			wantErr:  true,
			errField: "dates",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.meeting.Validate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

// TestMeeting_FindByUUID tests the FindByUUID function of the Meeting model
func (ms *ModelSuite) TestMeeting_FindByUUID() {
	t := ms.T()

	uf := createUserFixtures(ms.DB, 2)
	meetings := createMeetingFixtures_FindByUUID(ms, t, uf.Users)

	tests := []struct {
		name    string
		uuid    string
		want    Meeting
		wantErr bool
	}{
		{name: "good", uuid: meetings[0].UUID.String(), want: meetings[0]},
		{name: "blank uuid", uuid: "", wantErr: true},
		{name: "wrong uuid", uuid: domain.GetUUID().String(), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var meeting Meeting
			err := meeting.FindByUUID(test.uuid)
			if test.wantErr {
				ms.Error(err, "FindByUUID() did not return expected error")
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(test.want.UUID, meeting.UUID, "incorrect uuid")
		})
	}
}

// TestMeeting_FindOnDate tests the FindOnDate function of the Meeting model
func (ms *ModelSuite) TestMeeting_FindOnDate() {
	t := ms.T()

	meetings := createMeetingFixtures_FindByTime(ms)

	nearFuture := time.Time(meetings[1].EndDate) // Also the start date of meetings[2]
	farFuture := time.Time(meetings[3].EndDate).Add(domain.DurationDay * 2)

	tests := []struct {
		name    string
		want    []string
		testNow time.Time
	}{
		{name: "one for actual now", testNow: time.Now(), want: []string{meetings[2].Name}},
		{
			name: "two for now in near future", testNow: nearFuture,
			want: []string{meetings[1].Name, meetings[2].Name},
		},
		{name: "empty for now in far future", testNow: farFuture, want: []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var meetings Meetings
			err := meetings.FindOnDate(test.testNow)
			ms.NoError(err, "unexpected error")

			mNames := make([]string, len(meetings))
			for i, m := range meetings {
				mNames[i] = m.Name
			}

			ms.Equal(test.want, mNames, "incorrect list of future meetings")
		})
	}
}

func getMeetingNames(meetings Meetings) []string {
	mNames := make([]string, len(meetings))
	for i, m := range meetings {
		mNames[i] = m.Name
	}

	return mNames
}

// TestMeeting_FindOnOrAfterDate tests the FindOnOrAfterDate function of the Meeting model
func (ms *ModelSuite) TestMeeting_FindOnOrAfterDate() {
	t := ms.T()

	meetings := createMeetingFixtures_FindByTime(ms)

	// Two days after meetings[2] ends
	nearFuture := time.Time(meetings[2].EndDate).Add(domain.DurationDay * 2)

	// Two days after meetings[3] ends
	farFuture := time.Time(meetings[3].EndDate).Add(domain.DurationDay * 2)

	tests := []struct {
		name    string
		want    []string
		testNow time.Time
	}{
		{name: "two for actual now", testNow: time.Now(), want: []string{meetings[2].Name, meetings[3].Name}},
		{name: "one for now in future", testNow: nearFuture, want: []string{meetings[3].Name}},
		{name: "empty for now in far future", testNow: farFuture, want: []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var meetings Meetings
			err := meetings.FindOnOrAfterDate(test.testNow)
			ms.NoError(err, "unexpected error")

			mNames := getMeetingNames(meetings)
			ms.Equal(test.want, mNames, "incorrect list of future meetings")
		})
	}
}

// TestMeeting_FindAfterDate tests the FindAfterDate function of the Meeting model
func (ms *ModelSuite) TestMeeting_FindAfterDate() {
	t := ms.T()

	meetings := createMeetingFixtures_FindByTime(ms)

	// Two days before meetings[2] starts
	nearPast := time.Time(meetings[2].StartDate).Add(-domain.DurationDay * 2)

	// Two days after meetings[2] ends
	nearFuture := time.Time(meetings[2].EndDate).Add(domain.DurationDay * 2)

	tests := []struct {
		name    string
		want    []string
		testNow time.Time
	}{
		{name: "two for now in past", testNow: nearPast, want: []string{meetings[2].Name, meetings[3].Name}},
		{name: "one for actual now", testNow: time.Now(), want: []string{meetings[3].Name}},
		{name: "empty for now in near future", testNow: nearFuture, want: []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var meetings Meetings
			err := meetings.FindAfterDate(test.testNow)
			ms.NoError(err, "unexpected error")

			mNames := getMeetingNames(meetings)
			ms.Equal(test.want, mNames, "incorrect list of future meetings")
		})
	}
}

// TestMeeting_FindRecent tests the FindRecent function of the Meeting model
func (ms *ModelSuite) TestMeeting_FindRecent() {
	t := ms.T()

	meetings := createMeetingFixtures_FindByTime(ms)

	// Two days after meetings[2] ends
	nearFuture := time.Time(meetings[2].EndDate).Add(domain.DurationDay * 2)

	// Five weeks after meetings[3] ends
	farFuture := time.Time(meetings[3].EndDate).Add(domain.DurationWeek * 5)

	tests := []struct {
		name    string
		want    []string
		testNow time.Time
	}{
		{name: "one for actual now", testNow: time.Now(), want: []string{meetings[1].Name}},
		{name: "two for now in near future", testNow: nearFuture, want: []string{meetings[1].Name, meetings[2].Name}},
		{name: "empty far now in far future", testNow: farFuture, want: []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var meetings Meetings
			err := meetings.FindRecent(test.testNow)
			ms.NoError(err, "unexpected error")

			mNames := getMeetingNames(meetings)
			ms.Equal(test.want, mNames, "incorrect list of future meetings")
		})
	}
}

// TestMeeting_FindByUUID tests the FindByUUID function of the Meeting model
func (ms *ModelSuite) TestMeeting_FindByInviteCode() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		code    string
		want    Meeting
		wantErr bool
	}{
		{name: "good", code: f.Meetings[0].InviteCode.UUID.String(), want: f.Meetings[0]},
		{name: "blank uuid", code: "", wantErr: true},
		{name: "wrong uuid", code: domain.GetUUID().String(), wantErr: true},
	}
	for _, test := range tests {
		ms.T().Run(test.name, func(t *testing.T) {
			var meeting Meeting
			err := meeting.FindByInviteCode(test.code)
			if test.wantErr {
				ms.Error(err, "FindByInviteCode() did not return expected error")
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(test.want.UUID, meeting.UUID, "incorrect uuid")
		})
	}
}

func (ms *ModelSuite) TestMeeting_ImageFile() {
	user := User{}
	createFixture(ms, &user)

	location := Location{}
	createFixture(ms, &location)

	meeting := Meeting{
		Name:        "a meeting",
		UUID:        domain.GetUUID(),
		CreatedByID: user.ID,
		LocationID:  location.ID,
		EndDate:     time.Now(),
		StartDate:   time.Now(),
	}
	createFixture(ms, &meeting)

	f, err := meeting.ImageFile()
	ms.NoError(err, "unexpected error from Meeting.ImageFile()")
	ms.Nil(f, "expected nil returned from Meeting.ImageFile()")

	imageFixture := createFileFixture()

	attachedFile, err := meeting.SetImageFile(imageFixture.UUID.String())
	ms.NoError(err)

	if got, err := meeting.ImageFile(); err == nil {
		ms.Equal(attachedFile.UUID.String(), got.UUID.String())
		ms.True(got.URLExpiration.After(time.Now().Add(time.Minute)))
		ms.Equal(imageFixture.Name, got.Name)
	} else {
		ms.Fail("meeting.GetImage failed, %s", err)
	}
}

func (ms *ModelSuite) TestMeeting_GetCreator() {
	uf := createUserFixtures(ms.DB, 1)
	user := uf.Users[0]

	location := Location{}
	createFixture(ms, &location)

	meeting := Meeting{CreatedByID: user.ID, Name: "name", LocationID: location.ID}
	createFixture(ms, &meeting)

	creator, err := meeting.GetCreator()
	ms.NoError(err, "unexpected error from meeting.GetCreator()")
	ms.Equal(user.Nickname, creator.Nickname, "incorrect user/creator of meeting")
}

func (ms *ModelSuite) TestMeeting_GetSetLocation() {
	uf := createUserFixtures(ms.DB, 1)
	user := uf.Users[0]

	locations := Locations{
		{
			Description: "a place",
			Country:     "XY",
			Latitude:    nulls.NewFloat64(1.1),
			Longitude:   nulls.NewFloat64(2.2),
		},
		{
			Description: "another place",
			Country:     "AB",
			Latitude:    nulls.Float64{},
			Longitude:   nulls.Float64{},
		},
	}
	createFixture(ms, &locations[0]) // only save the first record for now

	meeting := Meeting{CreatedByID: user.ID, Name: "name", LocationID: locations[0].ID}
	createFixture(ms, &meeting)

	err := meeting.SetLocation(locations[1])
	ms.NoError(err, "unexpected error from meeting.SetLocation()")

	locationFromDB, err := meeting.GetLocation()
	ms.NoError(err, "unexpected error from meeting.GetLocation()")
	locations[1].ID = locationFromDB.ID
	ms.Equal(locations[1], locationFromDB, "location data doesn't match after update")

	// These are redundant checks, but here to document the fact that a null overwrites previous data.
	ms.False(locationFromDB.Latitude.Valid)
	ms.False(locationFromDB.Longitude.Valid)
}

func (ms *ModelSuite) TestMeeting_CanUpdate() {
	f := createMeetingFixtures_CanUpdate(ms)

	mtgUser := f.Users[0]
	superUser := f.Users[1]
	salesUser := f.Users[2]
	adminUser := f.Users[3]
	otherUser := f.Users[4]

	mtg := f.Meetings[0]

	ms.True(mtg.CanUpdate(mtgUser), "meeting creator should be authorized")
	ms.True(mtg.CanUpdate(superUser), "super admin should be authorized")
	ms.True(mtg.CanUpdate(salesUser), "sales admin should be authorized")
	ms.True(mtg.CanUpdate(adminUser), "admin should be authorized")
	ms.False(mtg.CanUpdate(otherUser), "normal user (non meeting creator) should NOT be authorized")
}

func (ms *ModelSuite) TestMeeting_GetRequests() {
	meetings := createMeetingFixtures(ms.DB, 2).Meetings

	requests := createRequestFixtures(ms.DB, 3, false)
	requests[0].MeetingID = nulls.NewInt(meetings[1].ID)
	requests[1].MeetingID = nulls.NewInt(meetings[1].ID)
	ms.NoError(ms.DB.Update(&requests))

	tests := []struct {
		name    string
		meeting Meeting
		wantIDs []int
		wantErr string
	}{
		{
			name:    "none",
			meeting: meetings[0],
			wantIDs: []int{},
		},
		{
			name:    "two",
			meeting: meetings[1],
			wantIDs: []int{requests[1].ID, requests[0].ID},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.meeting.Requests()
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")

			pIDs := make([]int, len(got))
			for i, p := range got {
				pIDs[i] = p.ID
			}

			ms.Equal(tt.wantIDs, pIDs)
		})
	}
}

func (ms *ModelSuite) TestMeeting_Invites() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name       string
		user       User
		meeting    Meeting
		wantEmails []string
		wantErr    string
	}{
		{
			name:       "creator",
			user:       f.Users[0],
			meeting:    f.Meetings[0],
			wantEmails: []string{f.MeetingInvites[0].Email, f.MeetingInvites[1].Email},
		},
		{
			name:       "organizer",
			user:       f.Users[1],
			meeting:    f.Meetings[0],
			wantEmails: []string{f.MeetingInvites[0].Email, f.MeetingInvites[1].Email},
		},
		{
			name:       "invitee",
			user:       f.Users[2],
			meeting:    f.Meetings[0],
			wantEmails: []string{},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.meeting.Invites(createTestContext(tt.user))
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")

			ids := make([]string, len(got))
			for i, invite := range got {
				ids[i] = invite.Email
				ms.Equal(tt.meeting.ID, invite.MeetingID, "wrong meeting ID in invite")
			}

			ms.Equal(tt.wantEmails, ids)
		})
	}
}

func (ms *ModelSuite) TestMeeting_Participants() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		user    User
		meeting Meeting
		wantIDs []int
		wantErr string
	}{
		{
			name:    "creator",
			user:    f.Users[0],
			meeting: f.Meetings[0],
			wantIDs: []int{f.Users[1].ID, f.Users[2].ID, f.Users[3].ID},
		},
		{
			name:    "organizer",
			user:    f.Users[1],
			meeting: f.Meetings[0],
			wantIDs: []int{f.Users[1].ID, f.Users[2].ID, f.Users[3].ID},
		},
		{
			name:    "participant",
			user:    f.Users[2],
			meeting: f.Meetings[0],
			wantIDs: []int{f.Users[2].ID},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.meeting.Participants(createTestContext(tt.user))
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")

			ids := make([]int, len(got))
			for i, p := range got {
				ids[i] = p.UserID
				ms.Equal(tt.meeting.ID, p.MeetingID, "wrong meeting ID in participant")
			}

			ms.Equal(tt.wantIDs, ids)
		})
	}
}

func (ms *ModelSuite) TestMeeting_Organizers() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		user    User
		meeting Meeting
		wantIDs []int
		wantErr string
	}{
		{
			name:    "creator",
			user:    f.Users[0],
			meeting: f.Meetings[0],
			wantIDs: []int{f.Users[1].ID},
		},
		{
			name:    "organizer",
			user:    f.Users[1],
			meeting: f.Meetings[0],
			wantIDs: []int{f.Users[1].ID},
		},
		{
			name:    "participant",
			user:    f.Users[2],
			meeting: f.Meetings[0],
			wantIDs: []int{f.Users[1].ID},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.meeting.Organizers(createTestContext(tt.user))
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")

			ids := make([]int, len(got))
			for i, u := range got {
				ids[i] = u.ID
				ms.Equal("", u.LastName, "organizer's name was not omitted")
			}

			ms.Equal(tt.wantIDs, ids)
		})
	}
}

func (ms *ModelSuite) TestMeeting_RemoveInvite() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name             string
		user             User
		meeting          Meeting
		email            string
		remainingInvites []string
		wantErr          string
	}{
		{
			name:    "wrong email",
			user:    f.Users[0],
			meeting: f.Meetings[0],
			email:   "not-there@example.com",
			wantErr: "no rows",
		},
		{
			name:             "one remaining",
			user:             f.Users[0],
			meeting:          f.Meetings[0],
			email:            f.MeetingInvites[0].Email,
			remainingInvites: []string{f.MeetingInvites[1].Email},
		},
		{
			name:             "none remaining",
			user:             f.Users[0],
			meeting:          f.Meetings[0],
			email:            f.MeetingInvites[1].Email,
			remainingInvites: []string{},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			// setup
			ctx := createTestContext(tt.user)

			// execute
			err := tt.meeting.RemoveInvite(ctx, tt.email)

			// verify
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")

			remaining, err := tt.meeting.Invites(ctx)
			ms.NoError(err)

			emails := make([]string, len(remaining))
			for i, m := range remaining {
				emails[i] = m.Email
			}

			ms.Equal(tt.remainingInvites, emails)

			// teardown
		})
	}
}

func (ms *ModelSuite) TestMeeting_RemoveParticipant() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name                  string
		testUser              User
		meeting               Meeting
		user                  User
		remainingParticipants []int
		wantErr               string
	}{
		{
			name:     "user not a participant",
			testUser: f.Users[0],
			meeting:  f.Meetings[0],
			user:     f.Users[0],
			wantErr:  "no rows",
		},
		{
			name:                  "good",
			testUser:              f.Users[0],
			meeting:               f.Meetings[0],
			user:                  f.Users[1],
			remainingParticipants: []int{f.MeetingParticipants[1].ID, f.MeetingParticipants[2].ID},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			// setup
			ctx := createTestContext(tt.testUser)

			// execute
			err := tt.meeting.RemoveParticipant(ctx, tt.user.UUID.String())

			// verify
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")

			remaining, err := tt.meeting.Participants(ctx)
			ms.NoError(err)

			ids := make([]int, len(remaining))
			for i, m := range remaining {
				ids[i] = m.ID
			}

			ms.Equal(tt.remainingParticipants, ids)

			// teardown
		})
	}
}

func (ms *ModelSuite) TestMeeting_isCodeValid() {
	code := domain.GetUUID()

	tests := []struct {
		name    string
		meeting Meeting
		code    string
		want    bool
	}{
		{
			name:    "yes",
			meeting: Meeting{InviteCode: nulls.NewUUID(code)},
			code:    code.String(),
			want:    true,
		},
		{
			name:    "wrong code",
			meeting: Meeting{InviteCode: nulls.NewUUID(domain.GetUUID())},
			code:    code.String(),
			want:    false,
		},
		{
			name:    "null-value code",
			meeting: Meeting{InviteCode: nulls.NewUUID(uuid.Nil)},
			code:    code.String(),
			want:    false,
		},
		{
			name:    "invalid code",
			meeting: Meeting{InviteCode: nulls.UUID{}},
			code:    code.String(),
			want:    false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			ms.Equal(tt.want, tt.meeting.IsCodeValid(tt.code), "IsCodeValid returned incorrect result")
		})
	}
}

func (ms *ModelSuite) TestMeeting_isOrganizer() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		user    User
		meeting Meeting
		want    bool
	}{
		{
			name:    "creator",
			user:    f.Users[0],
			meeting: f.Meetings[0],
			want:    false,
		},
		{
			name:    "organizer",
			user:    f.Users[1],
			meeting: f.Meetings[0],
			want:    true,
		},
		{
			name:    "participant",
			user:    f.Users[2],
			meeting: f.Meetings[0],
			want:    false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got := tt.meeting.isOrganizer(createTestContext(tt.user), tt.user.ID)
			ms.Equal(tt.want, got)
		})
	}
}

func (ms *ModelSuite) TestMeetings_FindByIDs() {
	t := ms.T()

	f := createMeetingFixtures(ms.DB, 3)
	meetings := f.Meetings

	tests := []struct {
		name string
		ids  []int
		want []string
	}{
		{
			name: "good",
			ids:  []int{meetings[0].ID, meetings[1].ID, meetings[0].ID},
			want: []string{meetings[0].Name, meetings[1].Name},
		},
		{
			name: "missing",
			ids:  []int{99999},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Meetings
			err := m.FindByIDs(tt.ids)
			ms.NoError(err)

			got := make([]string, len(m))
			for i, mm := range m {
				got[i] = mm.Name
			}
			ms.Equal(tt.want, got, "incorrect meeting names")
		})
	}
}
