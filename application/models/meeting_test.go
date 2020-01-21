package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate"

	"github.com/silinternational/wecarry-api/domain"
)

type meetingFixtures struct {
	Meetings
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
		{name: "two for now in near future", testNow: nearFuture,
			want: []string{meetings[1].Name, meetings[2].Name}},
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

// TestMeeting_AttachImage_GetImage tests the AttachImage and GetImage methods of models.Meeting
func (ms *ModelSuite) TestMeeting_AttachImage_GetImage() {
	t := ms.T()

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

	var imageFixture File
	const filename = "photo.gif"
	ms.Nil(imageFixture.Store(filename, []byte("GIF89a")), "failed to create file fixture")

	attachedFile, err := meeting.AttachImage(imageFixture.UUID.String())
	if err != nil {
		t.Errorf("failed to attach image to meeting, %s", err)
	} else {
		ms.Equal(filename, attachedFile.Name)
		ms.True(attachedFile.ID != 0)
		ms.True(attachedFile.UUID.Version() != 0)
	}

	if err := DB.Load(&meeting); err != nil {
		t.Errorf("failed to load image relation for test meeting, %s", err)
	}

	ms.Equal(filename, meeting.ImageFile.Name)

	if got, err := meeting.GetImage(); err == nil {
		ms.Equal(attachedFile.UUID.String(), got.UUID.String())
		ms.True(got.URLExpiration.After(time.Now().Add(time.Minute)))
		ms.Equal(filename, got.Name)
	} else {
		ms.Fail("meeting.GetImagefailed, %s", err)
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
