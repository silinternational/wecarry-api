package models

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate/v3"
)

func (ms *ModelSuite) TestRequest_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		request  Request
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr: false,
		},
		{
			name: "missing created_by",
			request: Request{
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "created_by",
		},
		{
			name: "missing organization_id",
			request: Request{
				CreatedByID: 1,
				Title:       "A Request",
				Size:        RequestSizeMedium,
				Status:      RequestStatusOpen,
				UUID:        domain.GetUUID(),
			},
			wantErr:  true,
			errField: "organization_id",
		},
		{
			name: "missing title",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Size:           RequestSizeMedium,
				Status:         RequestStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "missing size",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Status:         RequestStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "size",
		},
		{
			name: "missing status",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "missing uuid",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusOpen,
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "bad neededBefore (today)",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				NeededBefore:   nulls.NewTime(time.Now()),
				Size:           RequestSizeMedium,
				Status:         RequestStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "needed_before",
		},
		{
			name: "good neededBefore (tomorrow)",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				NeededBefore:   nulls.NewTime(time.Now().Add(domain.DurationDay)),
				Size:           RequestSizeMedium,
				Status:         RequestStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.request.Validate(DB)
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

func (ms *ModelSuite) TestRequest_ValidateCreate() {
	t := ms.T()

	tests := []struct {
		name     string
		request  Request
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "good - open",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr: false,
		},
		{
			name: "bad status - accepted",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusAccepted,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - delivered",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusDelivered,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - received",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusReceived,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - completed",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusCompleted,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - removed",
			request: Request{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           RequestSizeMedium,
				Status:         RequestStatusRemoved,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.request.ValidateCreate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_ValidateUpdate_OpenRequest() {
	t := ms.T()

	request := CreateFixturesValidateUpdate_RequestStatus(RequestStatusOpen, ms, t)

	tests := []struct {
		name    string
		request Request
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from open to open",
			request: Request{
				Status: RequestStatusOpen,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to accepted",
			request: Request{
				Title:  "New Title",
				Status: RequestStatusAccepted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to removed",
			request: Request{
				Status: RequestStatusRemoved,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to accepted",
			request: Request{
				Status: RequestStatusAccepted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from open to delivered",
			request: Request{
				Status: RequestStatusDelivered,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from open to received",
			request: Request{
				Status: RequestStatusReceived,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from open to completed",
			request: Request{
				Status: RequestStatusCompleted,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"
			vErr, _ := test.request.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_ValidateUpdate_AcceptedRequest() {
	t := ms.T()

	request := CreateFixturesValidateUpdate_RequestStatus(RequestStatusAccepted, ms, t)

	tests := []struct {
		name    string
		request Request
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from accepted to accepted",
			request: Request{
				Title:  "New Title",
				Status: RequestStatusAccepted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to open",
			request: Request{
				Status: RequestStatusOpen,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to delivered",
			request: Request{
				Status: RequestStatusDelivered,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to received",
			request: Request{
				Status: RequestStatusReceived,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to completed",
			request: Request{
				Status: RequestStatusCompleted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to removed",
			request: Request{
				Status: RequestStatusRemoved,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"
			vErr, _ := test.request.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_ValidateUpdate_DeliveredRequest() {
	t := ms.T()

	request := CreateFixturesValidateUpdate_RequestStatus(RequestStatusDelivered, ms, t)

	tests := []struct {
		name    string
		request Request
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from delivered to accepted",
			request: Request{
				Status: RequestStatusAccepted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from delivered to completed",
			request: Request{
				Status: RequestStatusCompleted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from delivered to open",
			request: Request{
				Status: RequestStatusOpen,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from delivered to received",
			request: Request{
				Title:  "New Title",
				Status: RequestStatusReceived,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from delivered to removed",
			request: Request{
				Status: RequestStatusRemoved,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"
			vErr, _ := test.request.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_ValidateUpdate_ReceivedRequest() {
	t := ms.T()

	request := CreateFixturesValidateUpdate_RequestStatus(RequestStatusReceived, ms, t)

	tests := []struct {
		name    string
		request Request
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from received to received",
			request: Request{
				Title:  "New Title",
				Status: RequestStatusReceived,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from received to accepted",
			request: Request{
				Status: RequestStatusAccepted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from received to completed",
			request: Request{
				Status: RequestStatusCompleted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from received to open",
			request: Request{
				Status: RequestStatusOpen,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from received to removed",
			request: Request{
				Status: RequestStatusRemoved,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"

			vErr, _ := test.request.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_ValidateUpdate_CompletedRequest() {
	t := ms.T()

	request := CreateFixturesValidateUpdate_RequestStatus(RequestStatusCompleted, ms, t)

	tests := []struct {
		name    string
		request Request
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from completed to completed",
			request: Request{
				Title:  "New Title",
				Status: RequestStatusCompleted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from completed to accepted",
			request: Request{
				Status: RequestStatusAccepted,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from completed to delivered",
			request: Request{
				Status: RequestStatusDelivered,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from completed to received",
			request: Request{
				Status: RequestStatusReceived,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from completed to open",
			request: Request{
				Status: RequestStatusOpen,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from completed to removed",
			request: Request{
				Status: RequestStatusRemoved,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"

			vErr, _ := test.request.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_ValidateUpdate_RemovedRequest() {
	t := ms.T()

	request := CreateFixturesValidateUpdate_RequestStatus(RequestStatusRemoved, ms, t)

	tests := []struct {
		name    string
		request Request
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from removed to removed",
			request: Request{
				Title:  "New Title",
				Status: RequestStatusRemoved,
				UUID:   request.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from removed to open",
			request: Request{
				Status: RequestStatusOpen,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from removed to accepted",
			request: Request{
				Status: RequestStatusAccepted,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from removed to delivered",
			request: Request{
				Status: RequestStatusDelivered,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from removed to received",
			request: Request{
				Status: RequestStatusReceived,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from removed to completed",
			request: Request{
				Status: RequestStatusCompleted,
				UUID:   request.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"

			vErr, _ := test.request.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_Create() {
	t := ms.T()
	f := createFixturesForTestRequestCreate(ms)

	tests := []struct {
		name    string
		request Request
		wantErr string
	}{
		{
			name:    "no uuid",
			request: f.Requests[0],
			wantErr: "",
		},
		{
			name:    "uuid given",
			request: f.Requests[1],
			wantErr: "",
		},
		{
			name:    "validation error",
			request: f.Requests[2],
			wantErr: "Title can not be blank.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Create(Ctx())
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ms.True(test.request.UUID.Version() != 0)
			var r Request
			ms.NoError(r.FindByID(ms.DB, test.request.ID))

			pHistories := RequestHistories{}
			err = ms.DB.Where("request_id = ?", r.ID).All(&pHistories)
			ms.NoError(err)

			ms.Equal(1, len(pHistories), "incorrect number of RequestHistories")
			ms.Equal(RequestStatusOpen, pHistories[0].Status, "incorrect status on RequestHistory")
		})
	}
}

func (ms *ModelSuite) TestRequest_Update() {
	t := ms.T()
	f := createFixturesForTestRequestUpdate(ms)

	tests := []struct {
		name    string
		request Request
		wantErr string
	}{
		{
			name:    "good",
			request: f.Requests[0],
			wantErr: "",
		},
		{
			name:    "validation error",
			request: f.Requests[1],
			wantErr: "Title can not be blank.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Update(Ctx())
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ms.True(test.request.UUID.Version() != 0)
			var r Request
			ms.NoError(r.FindByID(ms.DB, test.request.ID))
		})
	}
}

func (ms *ModelSuite) TestRequest_manageStatusTransition_forwardProgression() {
	t := ms.T()
	f := createFixturesForTestRequest_manageStatusTransition_forwardProgression(ms)

	tests := []struct {
		name            string
		request         Request
		newStatus       RequestStatus
		providerID      nulls.Int
		wantCompletedOn bool
	}{
		{
			name:      "open to open - no change",
			request:   f.Requests[0],
			newStatus: RequestStatusOpen,
		},
		{
			name:       "open to accepted - new history with provider",
			request:    f.Requests[0],
			newStatus:  RequestStatusAccepted,
			providerID: nulls.NewInt(f.Users[1].ID),
		},
		{
			name:            "accepted to completed - CompletedOn added",
			request:         f.Requests[2],
			newStatus:       RequestStatusCompleted,
			wantCompletedOn: true,
		},
		{
			name:            "delivered to completed - CompletedOn added",
			request:         f.Requests[3],
			newStatus:       RequestStatusCompleted,
			wantCompletedOn: true,
		},
		{
			name:       "open to accepted - new history with provider",
			request:    f.Requests[0],
			newStatus:  RequestStatusAccepted,
			providerID: nulls.NewInt(f.Users[1].ID),
		},
		{
			name:       "get error",
			request:    f.Requests[1],
			newStatus:  "BadStatus",
			providerID: f.Requests[1].ProviderID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.request.Status = test.newStatus
			test.request.ProviderID = test.providerID
			err := test.request.manageStatusTransition(ms.DB)
			ms.NoError(err)

			ph := RequestHistory{}
			err = ph.getLastForRequest(ms.DB, test.request)
			ms.NoError(err)

			ms.Equal(test.newStatus, ph.Status, "incorrect Status ")
			ms.Equal(test.providerID, ph.ProviderID, "incorrect ProviderID ")

			if test.wantCompletedOn {
				ms.True(test.request.CompletedOn.Valid, "expected a valid CompletedOn date")
			} else {
				ms.False(test.request.CompletedOn.Valid, "expected a null CompletedOn date")
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_manageStatusTransition_backwardProgression() {
	t := ms.T()
	f := createFixturesForTestRequest_manageStatusTransition_backwardProgression(ms)

	tests := []struct {
		name            string
		request         Request
		newStatus       RequestStatus
		providerID      nulls.Int
		wantCompletedOn bool
		wantErr         string
	}{
		{
			name:       "accepted to accepted - no change",
			request:    f.Requests[0],
			newStatus:  RequestStatusAccepted,
			providerID: f.Requests[0].ProviderID,
			wantErr:    "",
		},
		{
			name:       "accepted to open",
			request:    f.Requests[1],
			newStatus:  RequestStatusOpen,
			providerID: nulls.Int{},
			wantErr:    "",
		},
		{
			name:            "completed to accepted - CompletedOn Dropped",
			request:         f.Requests[2],
			newStatus:       RequestStatusAccepted,
			providerID:      f.Requests[2].ProviderID,
			wantCompletedOn: false,
		},
		{
			name:            "completed to delivered - CompletedOn Dropped",
			request:         f.Requests[3],
			newStatus:       RequestStatusDelivered,
			providerID:      f.Requests[3].ProviderID,
			wantCompletedOn: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.request.Status = test.newStatus
			test.request.ProviderID = test.providerID
			err := test.request.manageStatusTransition(ms.DB)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ph := RequestHistory{}
			err = ph.getLastForRequest(ms.DB, test.request)
			ms.NoError(err)

			ms.Equal(test.newStatus, ph.Status, "incorrect Status ")
			ms.Equal(test.providerID, ph.ProviderID, "incorrect ProviderID ")
			ms.Equal(test.wantCompletedOn, test.request.CompletedOn.Valid, "incorrect CompletedOn valuie")
		})
	}
}

func (ms *ModelSuite) TestRequest_FindByID() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	requests := createRequestFixtures(ms.DB, 2, false)

	tests := []struct {
		name          string
		id            int
		eagerFields   []string
		wantRequest   Request
		wantCreatedBy User
		wantProvider  User
		wantErr       bool
	}{
		{
			name:        "good with no related fields",
			id:          requests[0].ID,
			wantRequest: requests[0],
		},
		{
			name:          "good with a related field",
			id:            requests[1].ID,
			eagerFields:   []string{"CreatedBy"},
			wantRequest:   requests[1],
			wantCreatedBy: users[0],
		},
		{name: "zero ID", id: 0, wantErr: true},
		{name: "wrong id", id: 99999, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var request Request
			err := request.FindByID(ms.DB, test.id, test.eagerFields...)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantRequest.ID, request.ID, "bad request id")
				if test.wantCreatedBy.ID != 0 {
					ms.Equal(test.wantCreatedBy.ID, request.CreatedBy.ID, "bad request createdby id")
				}
				if test.wantProvider.ID != 0 {
					ms.Equal(test.wantProvider.ID, request.Provider.ID, "bad request provider id")
				}
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_FindByUUID() {
	t := ms.T()

	_ = createUserFixtures(ms.DB, 2)
	requests := createRequestFixtures(ms.DB, 1, false)

	tests := []struct {
		name    string
		uuid    string
		want    Request
		wantErr bool
	}{
		{name: "good", uuid: requests[0].UUID.String(), want: requests[0]},
		{name: "blank uuid", uuid: "", wantErr: true},
		{name: "wrong uuid", uuid: domain.GetUUID().String(), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var request Request
			err := request.FindByUUID(ms.DB, test.uuid)
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("FindByUUID() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("FindByUUID() error = %v", err)
				} else if request.UUID != test.want.UUID {
					t.Errorf("FindByUUID() got = %s, want %s", request.UUID, test.want.UUID)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_GetCreator() {
	t := ms.T()

	uf := createUserFixtures(ms.DB, 2)
	requests := createRequestFixtures(ms.DB, 1, false)

	tests := []struct {
		name    string
		request Request
		want    uuid.UUID
	}{
		{name: "good", request: requests[0], want: uf.Users[0].UUID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.request.GetCreator(ms.DB)
			if err != nil {
				t.Errorf("GetCreator() error = %v", err)
			} else if user.UUID != test.want {
				t.Errorf("GetCreator() got = %s, want %s", user.UUID, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_GetProvider() {
	t := ms.T()

	uf := createUserFixtures(ms.DB, 2)
	requests := createRequestFixtures(ms.DB, 2, false)
	requests[1].ProviderID = nulls.NewInt(uf.Users[1].ID)

	tests := []struct {
		name    string
		request Request
		want    *uuid.UUID
	}{
		{name: "good", request: requests[1], want: &uf.Users[1].UUID},
		{name: "nil", request: requests[0], want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.request.GetProvider(ms.DB)
			if err != nil {
				t.Errorf("GetProvider() error = %v", err)
			} else if test.want == nil {
				if user != nil {
					t.Errorf("expected nil, got %s", user.UUID.String())
				}
			} else if user == nil {
				t.Errorf("received nil, expected %v", test.want.String())
			} else if user.UUID != *test.want {
				t.Errorf("GetProvider() got = %s, want %s", user.UUID, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_GetStatusTransitions() {
	tests := []struct {
		name    string
		request Request
		user    User
		want    []StatusTransitionTarget
	}{
		{
			name:    "Open Request - Creator",
			request: Request{ID: 1, CreatedByID: 11, Status: RequestStatusOpen},
			user:    User{ID: 11},
			want: []StatusTransitionTarget{
				{Status: RequestStatusAccepted},
				{Status: RequestStatusRemoved},
			},
		},
		{
			name:    "Accepted Request - Creator",
			request: Request{ID: 1, CreatedByID: 11, Status: RequestStatusAccepted},
			user:    User{ID: 11},
			want: []StatusTransitionTarget{
				{Status: RequestStatusOpen, IsBackStep: true},
				{Status: RequestStatusReceived},
				{Status: RequestStatusCompleted},
				{Status: RequestStatusRemoved},
			},
		},
		{
			name:    "Delivered Request - Creator",
			request: Request{ID: 1, CreatedByID: 11, Status: RequestStatusDelivered},
			user:    User{ID: 11},
			want:    []StatusTransitionTarget{{Status: RequestStatusCompleted}},
		},
		{
			name:    "Completed Request - Creator",
			request: Request{ID: 1, CreatedByID: 11, Status: RequestStatusCompleted},
			user:    User{ID: 11},
			want: []StatusTransitionTarget{
				{Status: RequestStatusAccepted, IsBackStep: true},
				{Status: RequestStatusDelivered, IsBackStep: true},
			},
		},
		{
			name:    "Accepted Request - Provider",
			request: Request{ID: 1, ProviderID: nulls.NewInt(12), Status: RequestStatusAccepted},
			user:    User{ID: 12},
			want: []StatusTransitionTarget{
				{Status: RequestStatusDelivered, isProviderAction: true},
			},
		},
		{
			name:    "Delivered Request - Provider",
			request: Request{ID: 1, ProviderID: nulls.NewInt(12), Status: RequestStatusDelivered},
			user:    User{ID: 12},
			want: []StatusTransitionTarget{
				{Status: RequestStatusAccepted, IsBackStep: true, isProviderAction: true},
			},
		},
		{
			name:    "Completed Request - Provider",
			request: Request{ID: 1, ProviderID: nulls.NewInt(12), Status: RequestStatusCompleted},
			user:    User{ID: 12},
			want:    []StatusTransitionTarget{},
		},
		{
			name:    "Open Request - Not Creator Or Provider",
			request: Request{ID: 1, Status: RequestStatusOpen},
			user:    User{ID: 99}, want: []StatusTransitionTarget{},
		},
		{
			name:    "Accepted Request - Not Creator Or Provider",
			request: Request{ID: 1, Status: RequestStatusAccepted},
			user:    User{ID: 99}, want: []StatusTransitionTarget{},
		},
	}

	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.request.GetStatusTransitions(tt.user)
			ms.NoError(err)
			ms.Equal(tt.want, got, "incorrect status transitions")
		})
	}
}

func (ms *ModelSuite) TestRequest_GetPotentialProviderActions() {
	f := createUserFixtures(ms.DB, 3)
	users := f.Users
	requests := createRequestFixtures(ms.DB, 2, false)
	createPotentialProviderFixtures(ms.DB, 0, 2)

	acceptedRequest := requests[0]
	acceptedRequest.Status = RequestStatusAccepted // This doesn't change the request in the slice

	tests := []struct {
		name    string
		request Request
		user    User
		want    []string
	}{
		{
			name:    "Open Request - Creator",
			request: requests[1],
			user:    users[0],
			want:    []string{},
		},
		{
			name:    "Open Request with no offers - not Creator",
			request: requests[1],
			user:    users[1],
			want:    []string{RequestActionOffer},
		},
		{
			name:    "Open Request with offer - Offerer",
			request: requests[0],
			user:    users[1],
			want:    []string{RequestActionRetractOffer},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.request.GetPotentialProviderActions(Ctx(), tt.user)
			ms.NoError(err)
			ms.Equal(tt.want, got, "incorrect actions")
		})
	}
}

func (ms *ModelSuite) TestRequest_GetCurrentActions() {
	f := createUserFixtures(ms.DB, 3)
	users := f.Users
	requests := createRequestFixtures(ms.DB, 2, false)
	_ = createPotentialProviderFixtures(ms.DB, 0, 2)

	acceptedRequest := requests[0]
	acceptedRequest.Status = RequestStatusAccepted // This doesn't change the request in the slice
	acceptedRequest.ProviderID = nulls.NewInt(users[1].ID)

	// The rest of the scenarios are already tested elsewhere
	tests := []struct {
		name    string
		request Request
		user    User
		want    []string
	}{
		{
			name:    "Open Request with offers - Creator",
			request: requests[0],
			user:    users[0],
			want:    []string{RequestActionAccept, RequestActionRemove},
		},
		{
			name:    "Accepted Request - Creator",
			request: acceptedRequest,
			user:    users[0],
			want:    []string{RequestActionReopen, RequestActionReceive, RequestActionRemove},
		},
		{
			name:    "Open Request with no offers - not Creator",
			request: requests[1],
			user:    users[1],
			want:    []string{RequestActionOffer},
		},
		{
			name:    "Open Request with offer - not Creator",
			request: requests[0],
			user:    users[1],
			want:    []string{RequestActionRetractOffer},
		},
		{
			name:    "Accepted Request - Provider",
			request: acceptedRequest,
			user:    users[1],
			want:    []string{RequestActionDeliver},
		},
	}

	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.request.GetCurrentActions(Ctx(), tt.user)
			ms.NoError(err)
			ms.Equal(tt.want, got, "incorrect actions")
		})
	}
}

func (ms *ModelSuite) TestRequest_GetOrganization() {
	t := ms.T()

	_ = createUserFixtures(ms.DB, 2)
	requests := createRequestFixtures(ms.DB, 1, false)
	ms.NoError(ms.DB.Load(&requests, "Organization"))

	tests := []struct {
		name    string
		request Request
		want    uuid.UUID
	}{
		{name: "good", request: requests[0], want: requests[0].Organization.UUID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			org, err := test.request.GetOrganization(ms.DB)
			if err != nil {
				t.Errorf("GetOrganization() error = %v", err)
			} else if org.UUID != test.want {
				t.Errorf("GetOrganization() got = %s, want %s", org.UUID, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_GetThreads() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	requests := createRequestFixtures(ms.DB, 2, false)
	threadFixtures := CreateThreadFixtures(ms, requests[0])
	threads := threadFixtures.Threads

	tests := []struct {
		name    string
		request Request
		want    []uuid.UUID
	}{
		{name: "no threads", request: requests[1], want: []uuid.UUID{}},
		{name: "two threads", request: requests[0], want: []uuid.UUID{threads[1].UUID, threads[0].UUID}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.request.GetThreads(Ctx(), users[0])
			if err != nil {
				t.Errorf("GetThreads() error: %v", err)
			} else {
				ids := make([]uuid.UUID, len(got))
				for i := range got {
					ids[i] = got[i].UUID
				}
				if !reflect.DeepEqual(ids, test.want) {
					t.Errorf("GetThreads() got = %s, want %s", ids, test.want)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_AttachFile() {
	t := ms.T()

	user := User{}
	createFixture(ms, &user)

	organization := Organization{AuthConfig: "{}"}
	createFixture(ms, &organization)

	location := Location{}
	createFixture(ms, &location)

	request := Request{
		CreatedByID:    user.ID,
		OrganizationID: organization.ID,
		DestinationID:  location.ID,
	}
	createFixture(ms, &request)

	file := createFileFixture(ms.DB)

	if attachedFile, err := request.AttachFile(ms.DB, file.UUID.String()); err != nil {
		t.Errorf("failed to attach file to request, %s", err)
	} else {
		ms.Equal(file.Name, attachedFile.Name)
		ms.True(attachedFile.ID != 0)
		ms.True(attachedFile.UUID.Version() != 0)
	}

	if err := ms.DB.Load(&request); err != nil {
		t.Errorf("failed to load relations for test request, %s", err)
	}

	ms.Equal(1, len(request.Files))

	if err := ms.DB.Load(&(request.Files[0])); err != nil {
		t.Errorf("failed to load files relations for test request, %s", err)
	}

	ms.Equal(file.Name, request.Files[0].File.Name)
}

func (ms *ModelSuite) TestRequest_GetFiles() {
	f := CreateFixturesForRequestsGetFiles(ms)

	files, err := f.Requests[0].GetFiles(Ctx())
	ms.NoError(err, "failed to get files list for request, %s", err)

	ms.Equal(len(f.Files), len(files))

	// sort most recently updated first
	expectedFilenames := []string{
		f.Files[2].Name,
		f.Files[1].Name,
		f.Files[0].Name,
	}

	receivedFilenames := make([]string, len(files))
	for i := range files {
		receivedFilenames[i] = files[i].Name
	}

	ms.Equal(expectedFilenames, receivedFilenames, "incorrect list of files")
}

// TestRequest_GetPhoto tests the GetPhoto method of models.Request
func (ms *ModelSuite) TestRequest_GetPhotoID() {
	requests := createRequestFixtures(ms.DB, 1, false)
	request := requests[0]

	photoFixture := createFileFixture(ms.DB)

	attachedFile, err := request.AttachPhoto(Ctx(), photoFixture.UUID.String())
	ms.NoError(err, "failed to attach photo to request")
	ms.Equal(photoFixture.Name, attachedFile.Name)
	ms.True(attachedFile.ID != 0)
	ms.True(attachedFile.UUID.Version() != 0)

	ms.NoError(DB.Load(&request), "failed to load photo relation for test request")

	ms.Equal(photoFixture.Name, request.PhotoFile.Name)

	got, err := request.GetPhotoID(Ctx())
	ms.NoError(err, "unexpected error")
	attachedFileUUID := attachedFile.UUID.String()
	ms.Equal(&attachedFileUUID, got)
}

// TestRequest_GetPhoto tests the GetPhoto method of models.Request
func (ms *ModelSuite) TestRequest_GetPhoto() {
	requests := createRequestFixtures(ms.DB, 1, false)
	request := requests[0]

	photoFixture := createFileFixture(ms.DB)

	attachedFile, err := request.AttachPhoto(Ctx(), photoFixture.UUID.String())
	ms.NoError(err, "failed to attach photo to request")
	ms.Equal(photoFixture.Name, attachedFile.Name)
	ms.True(attachedFile.ID != 0)
	ms.True(attachedFile.UUID.Version() != 0)

	ms.NoError(DB.Load(&request), "failed to load photo relation for test request")

	ms.Equal(photoFixture.Name, request.PhotoFile.Name)

	if got, err := request.GetPhoto(ms.DB); err == nil {
		ms.Equal(attachedFile.UUID.String(), got.UUID.String())
		ms.Equal(attachedFile.Name, got.Name)
	} else {
		ms.Fail("request.GetPhoto failed, %s", err)
	}
}

//func (ms *ModelSuite) TestRequest_FindByUserAndUUID() {
//	t := ms.T()
//	f := createFixturesForRequestFindByUserAndUUID(ms)
//
//	tests := []struct {
//		name    string
//		user    User
//		request    Request
//		wantErr string
//	}{
//		{name: "user 0, request 0", user: f.Users[0], request: f.Requests[0]},
//		{name: "user 0, request 1", user: f.Users[0], request: f.Requests[1]},
//		{name: "user 0, request 2 Removed", user: f.Users[0], request: f.Requests[2], wantErr: "no rows in result set"},
//		{name: "user 1, request 0", user: f.Users[1], request: f.Requests[0]},
//		{name: "user 1, request 1", user: f.Users[1], request: f.Requests[1], wantErr: "no rows in result set"},
//		{name: "non-existent user", request: f.Requests[1], wantErr: "no rows in result set"},
//		{name: "non-existent request", user: f.Users[1], wantErr: "no rows in result set"},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			var request Request
//			err := request.FindByUserAndUUID(Ctx(), test.user, test.request.UUID.String())
//
//			if test.wantErr != "" {
//				ms.Error(err)
//				ms.Contains(err.Error(), test.wantErr, "unexpected error")
//				return
//			}
//
//			ms.NoError(err)
//			ms.Equal(test.request.ID, request.ID)
//		})
//	}
//}

func (ms *ModelSuite) TestRequest_GetSetDestination() {
	t := ms.T()

	user := User{UUID: domain.GetUUID(), Email: t.Name() + "_user@example.com", Nickname: t.Name() + "_User"}
	createFixture(ms, &user)

	organization := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &organization)

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

	request := Request{CreatedByID: user.ID, OrganizationID: organization.ID, DestinationID: locations[0].ID}
	createFixture(ms, &request)

	err := request.SetDestination(Ctx(), locations[1])
	ms.NoError(err, "unexpected error from request.SetDestination()")

	locationFromDB, err := request.GetDestination(ms.DB)
	ms.NoError(err, "unexpected error from request.GetDestination()")
	locations[1].ID = locationFromDB.ID
	ms.Equal(locations[1], *locationFromDB, "destination data doesn't match after update")

	// These are redundant checks, but here to document the fact that a null overwrites previous data.
	ms.False(locationFromDB.Latitude.Valid)
	ms.False(locationFromDB.Longitude.Valid)
}

func (ms *ModelSuite) TestRequest_Origin() {
	requests := createRequestFixtures(ms.DB, 1, false)
	request := requests[0]
	request.OriginID = nulls.Int{}
	ms.NoError(ms.DB.Save(&request))

	locationFixtures := Locations{
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

	err := request.SetOrigin(Ctx(), locationFixtures[0])
	ms.NoError(err, "unexpected error from request.SetOrigin()")

	locationFromDB, err := request.GetOrigin(ms.DB)
	ms.NoError(err, "unexpected error from request.GetOrigin()")

	locationFixtures[0].ID = locationFromDB.ID
	ms.Equal(locationFixtures[0], *locationFromDB, "origin data doesn't match new location")

	err = request.SetOrigin(Ctx(), locationFixtures[1])
	ms.NoError(err, "unexpected error from request.SetOrigin()")

	locationFromDB, err = request.GetOrigin(ms.DB)
	ms.NoError(err, "unexpected error from request.GetOrigin()")
	ms.Equal(locationFixtures[0].ID, locationFromDB.ID,
		"Location ID doesn't match -- location record was probably not reused")

	locationFixtures[1].ID = locationFromDB.ID
	ms.Equal(locationFixtures[1], *locationFromDB, "origin data doesn't match after update")

	// These are redundant checks, but here to document the fact that a null overwrites previous data.
	ms.False(locationFromDB.Latitude.Valid)
	ms.False(locationFromDB.Longitude.Valid)

	ms.NoError(request.RemoveOrigin(Ctx()))
	ms.False(request.OriginID.Valid, "expected the origin to have been removed")
	err = ms.DB.Find(locationFromDB, locationFromDB.ID)
	ms.Error(err, "expected error when looking for removed origin")
	ms.False(domain.IsOtherThanNoRows(err), "unexpected error type finding old origin, "+err.Error())
}

func (ms *ModelSuite) TestRequest_NewWithUser() {
	t := ms.T()
	user := createUserFixtures(ms.DB, 1).Users[0]

	tests := []struct {
		name              string
		wantRequestStatus RequestStatus
		wantProviderID    int
		wantErr           bool
	}{
		{name: "Good Request", wantRequestStatus: RequestStatusOpen},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var request Request
			err := request.NewWithUser(user)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(user.ID, request.CreatedByID)
				ms.Equal(test.wantRequestStatus, request.Status)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_SetProviderWithStatus() {
	t := ms.T()
	user := createUserFixtures(ms.DB, 1).Users[0]

	tests := []struct {
		name           string
		status         RequestStatus
		wantProviderID nulls.Int
	}{
		{name: "Accepted Request", status: RequestStatusAccepted, wantProviderID: nulls.NewInt(user.ID)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var request Request
			userID := user.UUID.String()
			err := request.SetProviderWithStatus(Ctx(), test.status, &userID)
			ms.NoError(err)

			ms.Equal(test.wantProviderID, request.ProviderID)
			ms.Equal(test.status, request.Status)
		})
	}
}

func (ms *ModelSuite) TestRequests_FindByUser() {
	t := ms.T()

	f := CreateFixtures_Requests_FindByUser(ms)

	var requestZeroDestination Location
	ms.NoError(ms.DB.Find(&requestZeroDestination, f.Requests[0].DestinationID))
	var requestOneOrigin Location
	ms.NoError(ms.DB.Find(&requestOneOrigin, f.Requests[1].OriginID))

	tests := []struct {
		name           string
		user           User
		dest           *Location
		orig           *Location
		requestID      *int
		wantRequestIDs []int
		wantErr        bool
	}{
		{
			name: "user 0", user: f.Users[0],
			wantRequestIDs: []int{f.Requests[6].ID, f.Requests[5].ID, f.Requests[4].ID, f.Requests[1].ID, f.Requests[0].ID},
		},
		{name: "user 1", user: f.Users[1], wantRequestIDs: []int{f.Requests[5].ID, f.Requests[4].ID, f.Requests[0].ID}},
		{name: "user 2", user: f.Users[2], wantRequestIDs: []int{f.Requests[7].ID, f.Requests[6].ID, f.Requests[5].ID}},
		{name: "user 3", user: f.Users[3], wantRequestIDs: []int{f.Requests[6].ID, f.Requests[5].ID, f.Requests[1].ID}},
		{name: "non-existent user", user: User{}, wantErr: true},
		{name: "destination", user: f.Users[0], dest: &requestZeroDestination, wantRequestIDs: []int{f.Requests[0].ID}},
		{name: "origin", user: f.Users[0], orig: &requestOneOrigin, wantRequestIDs: []int{f.Requests[1].ID}},
		{name: "user 0, request 1 (visible)", user: f.Users[0], requestID: &f.Requests[1].ID, wantRequestIDs: []int{f.Requests[1].ID}},
		{name: "user 0, request 2 (not visible)", user: f.Users[0], requestID: &f.Requests[2].ID, wantRequestIDs: []int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := Requests{}
			filter := RequestFilterParams{
				Destination: test.dest,
				Origin:      test.orig,
				RequestID:   test.requestID,
			}
			err := requests.FindByUser(Ctx(), test.user, filter)

			if test.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			requestIDs := make([]int, len(requests))
			for i := range requests {
				requestIDs[i] = requests[i].ID
			}
			ms.Equal(test.wantRequestIDs, requestIDs)
		})
	}
}

func (ms *ModelSuite) TestRequests_GetPotentialProviders() {
	t := ms.T()

	f := createPotentialProvidersFixtures(ms)
	users := f.Users
	requests := f.Requests
	pps := f.PotentialProviders

	tests := []struct {
		name      string
		request   Request
		user      User
		wantPPIDs []int
	}{
		{
			name: "pps for first request by requester", request: requests[0], user: users[0],
			wantPPIDs: []int{pps[0].UserID, pps[1].UserID, pps[2].UserID},
		},
		{
			name: "pps for first request by one of the potential providers", request: requests[0], user: users[1],
			wantPPIDs: []int{pps[0].UserID},
		},
		{
			name: "pps for second request by a non potential provider", request: requests[1], user: users[1],
			wantPPIDs: []int{},
		},
		{name: "no pps for third request", request: requests[2], wantPPIDs: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.request
			pps, err := request.GetPotentialProviders(Ctx(), tt.user)
			ms.NoError(err, "unexpected error")

			ids := make([]int, len(pps))
			for i, pp := range pps {
				ids[i] = pp.ID
			}
			ms.Equal(tt.wantPPIDs, ids)
		})
	}
}

func (ms *ModelSuite) TestRequests_FindByUser_SearchText() {
	t := ms.T()
	f := createFixtures_Requests_FindByUser_SearchText(ms)

	tests := []struct {
		name           string
		user           User
		matchText      string
		wantRequestIDs []int
		wantErr        bool
	}{
		{
			name: "user 0 matching case request", user: f.Users[0], matchText: "Match",
			wantRequestIDs: []int{f.Requests[5].ID, f.Requests[1].ID, f.Requests[0].ID},
		},
		{
			name: "user 0 lower case request", user: f.Users[0], matchText: "match",
			wantRequestIDs: []int{f.Requests[5].ID, f.Requests[1].ID, f.Requests[0].ID},
		},
		{
			name: "user 1", user: f.Users[1], matchText: "Match",
			wantRequestIDs: []int{f.Requests[5].ID, f.Requests[1].ID},
		},
		{
			name: "non-existent user", user: User{}, matchText: "Match",
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := Requests{}
			err := requests.FindByUser(Ctx(), test.user, RequestFilterParams{SearchText: &test.matchText})

			if test.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			requestIDs := make([]int, len(requests))
			for i := range requests {
				requestIDs[i] = requests[i].ID
			}
			ms.Equal(test.wantRequestIDs, requestIDs)
		})
	}
}

func (ms *ModelSuite) TestRequest_IsEditable() {
	t := ms.T()

	f := CreateFixtures_Request_IsEditable(ms)

	tests := []struct {
		name    string
		user    User
		request Request
		want    bool
		wantErr bool
	}{
		{name: "user 0, request 0", user: f.Users[0], request: f.Requests[0], want: true},
		{name: "user 0, request 1", user: f.Users[0], request: f.Requests[1], want: false},
		{name: "user 1, request 0", user: f.Users[1], request: f.Requests[0], want: false},
		{name: "user 1, request 1", user: f.Users[1], request: f.Requests[1], want: false},
		{name: "non-existent user", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			editable, err := test.request.IsEditable(Ctx(), test.user)

			if test.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			ms.Equal(test.want, editable)
		})
	}
}

func (ms *ModelSuite) TestRequest_isRequestEditable() {
	t := ms.T()

	tests := []struct {
		status RequestStatus
		want   bool
	}{
		{status: RequestStatusOpen, want: true},
		{status: RequestStatusAccepted, want: true},
		{status: RequestStatusReceived, want: true},
		{status: RequestStatusDelivered, want: true},
		{status: RequestStatusCompleted, want: false},
		{status: RequestStatusRemoved, want: false},
		{status: RequestStatus(""), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			r := Request{Status: tt.status}
			if got := r.isRequestEditable(); got != tt.want {
				t.Errorf("isStatusEditable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_canUserChangeStatus() {
	t := ms.T()

	tests := []struct {
		name      string
		request   Request
		user      User
		newStatus RequestStatus
		want      bool
	}{
		{
			name:    "Creator",
			request: Request{CreatedByID: 1},
			user:    User{ID: 1},
			want:    true,
		},
		{
			name:    "SuperAdmin",
			request: Request{},
			user:    User{AdminRole: UserAdminRoleSuperAdmin},
			want:    true,
		},
		{
			name:      "Open",
			request:   Request{CreatedByID: 1},
			newStatus: RequestStatusOpen,
			want:      false,
		},
		{
			name:      "Open to Accepted",
			request:   Request{CreatedByID: 1, Status: RequestStatusOpen},
			user:      User{ID: 1},
			newStatus: RequestStatusAccepted,
			want:      true,
		},
		{
			name:      "Accepted",
			request:   Request{CreatedByID: 1},
			newStatus: RequestStatusAccepted,
			want:      false,
		},
		{
			name:      "Request Received",
			request:   Request{CreatedByID: 1},
			newStatus: RequestStatusReceived,
			want:      false,
		},
		{
			name:      "Request From Delivered to Accepted By Requester",
			newStatus: RequestStatusAccepted,
			request:   Request{CreatedByID: 1, ProviderID: nulls.NewInt(2), Status: RequestStatusDelivered},
			user:      User{ID: 1},
			want:      false,
		},
		{
			name:      "Request From Delivered to Accepted By Provider",
			newStatus: RequestStatusAccepted,
			request:   Request{CreatedByID: 1, ProviderID: nulls.NewInt(2), Status: RequestStatusDelivered},
			user:      User{ID: 2},
			want:      true,
		},
		{
			name:      "Request From Delivered to Accepted By non-Provider",
			newStatus: RequestStatusAccepted,
			request:   Request{CreatedByID: 1, ProviderID: nulls.NewInt(2), Status: RequestStatusDelivered},
			user:      User{ID: 3},
			want:      false,
		},
		{
			name:      "Request Delivered By Provider",
			newStatus: RequestStatusDelivered,
			request:   Request{CreatedByID: 1, ProviderID: nulls.NewInt(2)},
			user:      User{ID: 2},
			want:      true,
		},
		{
			name:      "Request Delivered By non-Provider",
			newStatus: RequestStatusDelivered,
			request:   Request{CreatedByID: 1, ProviderID: nulls.NewInt(2)},
			user:      User{ID: 3},
			want:      false,
		},
		{
			name:      "Completed",
			request:   Request{CreatedByID: 1},
			newStatus: RequestStatusCompleted,
			want:      false,
		},
		{
			name:      "Completed by Requester",
			request:   Request{CreatedByID: 1},
			newStatus: RequestStatusCompleted,
			user:      User{ID: 1},
			want:      true,
		},
		{
			name:      "Removed",
			request:   Request{CreatedByID: 1},
			newStatus: RequestStatusRemoved,
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.request.canUserChangeStatus(tt.user, tt.newStatus)
			ms.Equal(tt.want, got)
		})
	}
}

func (ms *ModelSuite) TestRequest_GetAudience() {
	t := ms.T()
	f := createFixturesForRequestGetAudience(ms)

	tests := []struct {
		name    string
		request Request
		want    []int
		wantErr string
	}{
		{
			name:    "basic",
			request: f.Requests[0],
			want:    []int{f.Users[0].ID, f.Users[1].ID},
		},
		{
			name:    "no users",
			request: f.Requests[1],
			want:    []int{},
		},
		{
			name:    "invalid request",
			request: Request{},
			want:    []int{},
			wantErr: "invalid request ID in GetAudience",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.request.GetAudience(ms.DB)
			if tt.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tt.wantErr)
				return
			}

			ms.NoError(err)

			ids := make([]int, len(got))
			for i := range got {
				ids[i] = got[i].ID
			}
			if !reflect.DeepEqual(ids, tt.want) {
				t.Errorf("GetAudience()\ngot = %v\nwant %v", ids, tt.want)
			}
		})
	}
}

func (ms *ModelSuite) TestRequest_Meeting() {
	t := ms.T()
	requests := createRequestFixtures(ms.DB, 2, false)
	meeting := Meeting{
		UUID:        domain.GetUUID(),
		Name:        "a meeting",
		CreatedByID: requests[0].CreatedByID,
		LocationID:  requests[0].DestinationID,
	}
	createFixture(ms, &meeting)
	requests[0].MeetingID = nulls.NewInt(meeting.ID)
	ms.NoError(ms.DB.Save(&requests[0]))

	tests := []struct {
		name    string
		request Request
		want    *uuid.UUID
	}{
		{
			name:    "has meeting",
			request: requests[0],
			want:    &meeting.UUID,
		},
		{
			name:    "no meeting",
			request: requests[1],
			want:    nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.request.Meeting(ms.DB)
			ms.NoError(err)
			if test.want == nil {
				ms.Nil(got)
				return
			}
			ms.Equal(*test.want, got.UUID)
		})
	}
}

func (ms *ModelSuite) TestRequest_DestroyPotentialProviders() {
	f := createPotentialProvidersFixtures(ms)
	requests := f.Requests
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		currentUser User
		request     Request
		status      RequestStatus
		wantIDs     []int
		wantErr     string
	}{
		{
			name:        "no change: wrong status",
			currentUser: users[0],
			request:     requests[1],
			status:      RequestStatusAccepted,
			wantIDs:     []int{pps[0].ID, pps[1].ID, pps[2].ID, pps[3].ID, pps[4].ID},
		},
		{
			name:        "good: Request Creator as current user",
			currentUser: users[0],
			request:     requests[0],
			status:      RequestStatusCompleted,
			wantIDs:     []int{pps[3].ID, pps[4].ID},
		},
		{
			name:        "bad: current user is potential provider but not Request Creator",
			currentUser: users[2],
			request:     requests[0],
			status:      RequestStatusCompleted,
			wantErr: fmt.Sprintf(`user %v has insufficient permissions to destroy PotentialProviders for Request %v`,
				users[2].ID, requests[0].ID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.DestroyPotentialProviders(Ctx(), test.status, test.currentUser)

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

func (ms *ModelSuite) TestRequest_IsVisible() {
	f := CreateFixtures_Requests_FindByUser(ms)

	tests := []struct {
		name    string
		user    User
		request Request
		want    bool
	}{
		{name: "request in same org", user: f.Users[0], request: f.Requests[0], want: true},
		{name: "COMPLETED request in same org", user: f.Users[0], request: f.Requests[2], want: false},
		{name: "REMOVED request in same org", user: f.Users[0], request: f.Requests[3], want: false},
		{name: "request visibility ALL in trusted org", user: f.Users[0], request: f.Requests[5], want: true},
		{name: "request visibility TRUSTED in trusted org", user: f.Users[0], request: f.Requests[6], want: true},
		{name: "request visibility SAME in trusted org", user: f.Users[0], request: f.Requests[7], want: false},
		{name: "bad user", user: User{}, request: f.Requests[5], want: false},
		{name: "bad request", user: f.Users[0], request: Request{}, want: false},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got := tt.request.IsVisible(createTestContext(tt.user), tt.user)
			ms.Equal(tt.want, got)
		})
	}
}
