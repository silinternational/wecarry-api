package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/validate"
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

// CreateFixtures_ThreadParticipant_SetLastViewedAt creates test fixtures for the ThreadParticipant_SetLastViewedAt test
func CreateFixtures_ThreadParticipant_SetLastViewedAt(ms *ModelSuite, t *testing.T) ThreadFixtures {
	ResetTables(t, ms.DB)

	org := Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(t, &org)

	users := Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(t, &(users[i]))
	}

	userOrgs := UserOrganizations{
		{OrganizationID: org.ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
	}
	for i := range userOrgs {
		createFixture(t, &(userOrgs[i]))
	}

	posts := Posts{
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeRequest,
			OrganizationID: org.ID,
			Status:         PostStatusOpen,
			Title:          "Maple Syrup",
			Size:           PostSizeMedium,
			Uuid:           domain.GetUuid(),
			NeededAfter:    time.Now(),
			NeededBefore:   time.Date(2029, time.August, 3, 0, 0, 0, 0, time.UTC),
			Category:       "Unknown",
		},
	}
	for i := range posts {
		createFixture(t, &(posts[i]))
	}

	threads := Threads{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(t, &(threads[i]))
	}

	threadParticipants := ThreadParticipants{
		{
			ThreadID:     threads[0].ID,
			UserID:       users[0].ID,
			LastViewedAt: time.Now().Add(-1 * time.Hour),
		},
	}
	for i := range threadParticipants {
		createFixture(t, &(threadParticipants[i]))
	}

	return ThreadFixtures{
		Users:              users,
		ThreadParticipants: threadParticipants,
	}
}

func (ms *ModelSuite) TestThreadParticipant_SetLastViewedAt() {
	t := ms.T()
	ResetTables(t, ms.DB)

	f := CreateFixtures_ThreadParticipant_SetLastViewedAt(ms, t)

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
			err := tp.SetLastViewedAt(test.lastViewedAt)
			if test.wantErr {
				ms.Error(err)
			} else {
				if err != nil {
					t.Errorf("SetLastViewedAt() returned an error: %v", err)
				} else {
					want := test.lastViewedAt.Add(-1 * time.Minute)
					ms.True(tp.LastViewedAt.After(want),
						fmt.Sprintf("time not correct, got %v, wanted afer %v", tp.LastViewedAt, want))
					want = test.lastViewedAt.Add(time.Minute)
					ms.True(tp.LastViewedAt.Before(want),
						fmt.Sprintf("time not correct, got %v, wanted before %v", tp.LastViewedAt, want))
				}
			}
		})
	}
}
