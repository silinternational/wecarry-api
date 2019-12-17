package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/validate"

	"github.com/silinternational/wecarry-api/domain"
)

// TestMeeting_Validate ensures errors are thrown for missing required fields
func (ms *ModelSuite) TestMeeting_Validate() {
	t := ms.T()
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
				StartDate:   time.Now(),
				EndDate:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing created_by",
			meeting: Meeting{
				UUID:       domain.GetUUID(),
				Name:       "A Meeting",
				LocationID: 1,
				StartDate:  time.Now(),
				EndDate:    time.Now(),
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
				StartDate:   time.Now(),
				EndDate:     time.Now(),
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
				EndDate:     time.Now(),
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
				StartDate:   time.Now(),
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
				StartDate:   time.Now(),
				EndDate:     time.Now(),
			},
			wantErr:  true,
			errField: "uuid",
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
	meetings := CreateMeetingFixtures(ms, t, uf.Users)

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

// TestMeeting_AttachImage_GetImage tests the AttachImage and GetImage methods of models.Meeting
func (ms *ModelSuite) TestMeeting_AttachImage_GetImage() {
	t := ms.T()

	user := User{}
	createFixture(ms, &user)

	location := Location{}
	createFixture(ms, &location)

	meeting := Meeting{
		CreatedByID: user.ID,
		LocationID:  location.ID,
	}
	createFixture(ms, &meeting)

	var imageFixture File
	const filename = "photo.gif"
	if err := imageFixture.Store(filename, []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
	}

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
