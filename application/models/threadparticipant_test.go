package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/validate/v3"
	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestThreadParticipant_Validate() {
	t := ms.T()
	tests := []struct {
		name              string
		threadParticipant ThreadParticipant
		want              *validate.Errors
		wantErr           bool
		errField          string
	}{
		{
			name: "minimum",
			threadParticipant: ThreadParticipant{
				ThreadID: 1,
				UserID:   1,
			},
			wantErr: false,
		},
		{
			name: "missing thread_id",
			threadParticipant: ThreadParticipant{
				UserID: 1,
			},
			wantErr:  true,
			errField: "thread_id",
		},
		{
			name: "missing user_id",
			threadParticipant: ThreadParticipant{
				ThreadID: 1,
			},
			wantErr:  true,
			errField: "user_id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.threadParticipant.Validate(DB)
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

// CreateFixtures_ThreadParticipant_UpdateLastViewedAt is used by
// TestThreadParticipant_UpdateLastViewedAt and TestThread_UpdateLastViewedAt
func CreateFixtures_ThreadParticipant_UpdateLastViewedAt(ms *ModelSuite, t *testing.T) ThreadFixtures {
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

	location := Location{}
	createFixture(ms, &location)

	requests := Requests{
		{
			CreatedByID:    users[0].ID,
			OrganizationID: org.ID,
			Status:         RequestStatusOpen,
			Title:          "Maple Syrup",
			Size:           RequestSizeMedium,
			UUID:           domain.GetUUID(),
			DestinationID:  location.ID,
		},
	}
	for i := range requests {
		createFixture(ms, &requests[i])
	}

	threads := Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
	}
	for i := range threads {
		createFixture(ms, &threads[i])
	}

	threadParticipants := ThreadParticipants{
		{
			ThreadID:     threads[0].ID,
			UserID:       users[0].ID,
			LastViewedAt: time.Now().Add(-1 * time.Hour),
		},
	}
	for i := range threadParticipants {
		createFixture(ms, &threadParticipants[i])
	}

	return ThreadFixtures{
		Users:              users,
		Threads:            threads,
		ThreadParticipants: threadParticipants,
	}
}

func (ms *ModelSuite) TestThreadParticipant_UpdateLastViewedAt() {
	t := ms.T()

	f := CreateFixtures_ThreadParticipant_UpdateLastViewedAt(ms, t)

	tests := []struct {
		name              string
		threadParticipant ThreadParticipant
		lastViewedAt      time.Time
		wantErr           bool
	}{
		{name: "good", threadParticipant: f.ThreadParticipants[0], lastViewedAt: time.Now()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tp := test.threadParticipant
			err := tp.UpdateLastViewedAt(test.lastViewedAt)

			// reload from database to ensure the new time was saved
			_ = ms.DB.Reload(&tp)

			if test.wantErr {
				ms.Error(err)
				return
			}

			if err != nil {
				t.Errorf("UpdateLastViewedAt() returned an error: %v", err)
				return
			}

			want := test.lastViewedAt.Add(-1 * time.Minute)
			ms.True(tp.LastViewedAt.After(want),
				fmt.Sprintf("time not correct, got %v, wanted afer %v", tp.LastViewedAt, want))
			want = test.lastViewedAt.Add(time.Minute)
			ms.True(tp.LastViewedAt.Before(want),
				fmt.Sprintf("time not correct, got %v, wanted before %v", tp.LastViewedAt, want))
		})
	}
}

func CreateFixtures_ThreadParticipant_FindByThreadIDAndUserID(ms *ModelSuite) ThreadFixtures {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	users := uf.Users

	location := Location{}
	createFixture(ms, &location)

	requests := Requests{
		{UUID: domain.GetUUID(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
	}
	for i := range requests {
		createFixture(ms, &requests[i])
	}

	threads := Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
	}
	for i := range threads {
		createFixture(ms, &threads[i])
	}

	threadParticipants := ThreadParticipants{
		{ThreadID: threads[0].ID, UserID: users[0].ID},
	}
	for i := range threadParticipants {
		createFixture(ms, &threadParticipants[i])
	}

	return ThreadFixtures{
		Users:              users,
		Threads:            threads,
		ThreadParticipants: threadParticipants,
	}
}

func (ms *ModelSuite) TestThreadParticipant_FindByThreadIDAndUserID() {
	t := ms.T()

	f := CreateFixtures_ThreadParticipant_FindByThreadIDAndUserID(ms)

	tests := []struct {
		name     string
		threadID int
		userID   int
		wantID   int
		wantErr  bool
	}{
		{name: "good", threadID: f.Threads[0].ID, userID: f.Users[0].ID, wantID: f.ThreadParticipants[0].ID},
		{name: "bad user ID", threadID: f.Threads[0].ID, userID: 0, wantErr: true},
		{name: "bad thread ID", threadID: 0, userID: f.Users[0].ID, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var tp ThreadParticipant
			err := tp.FindByThreadIDAndUserID(test.threadID, test.userID)

			if test.wantErr {
				ms.Error(err, "did not get an error from FindByThreadIDAndUserID")
				return
			}

			ms.NoError(err, "unexpected error from FindByThreadIDAndUserID")

			ms.Equal(test.wantID, tp.ID, "incorrect thread_participant ID returned")
		})
	}
}

// CreateFixtures_ThreadParticipant_UpdateLastNotifiedAt creates test fixtures for the
// ThreadParticipant_UpdateLastNotifiedAt test
func CreateFixtures_ThreadParticipant_UpdateLastNotifiedAt(ms *ModelSuite, t *testing.T) ThreadFixtures {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	users := uf.Users

	location := Location{}
	createFixture(ms, &location)

	requests := Requests{
		{UUID: domain.GetUUID(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
	}
	for i := range requests {
		createFixture(ms, &requests[i])
	}

	threads := Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
	}
	for i := range threads {
		createFixture(ms, &threads[i])
	}

	threadParticipants := ThreadParticipants{
		{
			ThreadID:       threads[0].ID,
			UserID:         users[0].ID,
			LastNotifiedAt: time.Now().Add(-1 * time.Hour),
		},
	}
	for i := range threadParticipants {
		createFixture(ms, &threadParticipants[i])
	}

	return ThreadFixtures{
		ThreadParticipants: threadParticipants,
	}
}

func (ms *ModelSuite) TestThreadParticipant_UpdateLastNotifiedAt() {
	t := ms.T()

	f := CreateFixtures_ThreadParticipant_UpdateLastNotifiedAt(ms, t)

	tests := []struct {
		name              string
		threadParticipant ThreadParticipant
		LastNotifiedAt    time.Time
		wantErr           bool
	}{
		{name: "now", threadParticipant: f.ThreadParticipants[0], LastNotifiedAt: time.Now()},
		{name: "future", threadParticipant: f.ThreadParticipants[0], LastNotifiedAt: time.Now().Add(time.Minute)},
		{name: "past", threadParticipant: f.ThreadParticipants[0], LastNotifiedAt: time.Now().Add(-1 * time.Minute)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tp := test.threadParticipant
			err := tp.UpdateLastNotifiedAt(test.LastNotifiedAt)

			// reload from database to ensure the new time was saved
			_ = ms.DB.Reload(&tp)

			if test.wantErr {
				ms.Error(err)
				return
			}

			if err != nil {
				t.Errorf("UpdateLastNotifiedAt() returned an error: %v", err)
				return
			}

			ms.WithinDuration(test.LastNotifiedAt, tp.LastNotifiedAt, time.Second,
				"time not correct, got %v, wanted %v", tp.LastNotifiedAt, test.LastNotifiedAt)
		})
	}
}
