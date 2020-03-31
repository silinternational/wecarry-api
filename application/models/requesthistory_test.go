package models

import (
	"bytes"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"os"
	"testing"
)

func (ms *ModelSuite) TestRequestHistory_Load() {
	t := ms.T()
	f := createFixturesForTestRequestHistory_Load(ms)

	tests := []struct {
		name           string
		requestHistory RequestHistory
		wantErr        string
		wantEmail      string
	}{
		{
			name:           "open",
			requestHistory: f.RequestHistories[0],
			wantEmail:      f.Users[0].Email,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.requestHistory.Load("Receiver")
			ms.NoError(err, "did not expect any error")

			ms.Equal(test.wantEmail, test.requestHistory.Receiver.Email, "incorrect Receiver email")

		})
	}
}

func (ms *ModelSuite) TestRequestHistory_pop() {
	t := ms.T()
	f := createFixturesForTestRequestHistory_pop(ms)

	tests := []struct {
		name          string
		request       Request
		newStatus     RequestStatus
		currentStatus RequestStatus
		providerID    int
		wantErr       string
		wantLog       string
		want          RequestHistories
	}{
		{
			name:      "null to open - log error",
			request:   f.Requests[1],
			newStatus: RequestStatusOpen,
			wantLog:   "None Found",
			want:      RequestHistories{},
		},
		{
			name:          "from accepted back to open",
			request:       f.Requests[0],
			newStatus:     RequestStatusOpen,
			currentStatus: RequestStatusAccepted,
			want:          RequestHistories{f.RequestHistories[0]},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var buf bytes.Buffer
			domain.ErrLogger.SetOutput(&buf)

			defer func() {
				domain.Logger.SetOutput(os.Stdout)
			}()

			test.request.Status = test.newStatus
			var pHistory RequestHistory

			err := pHistory.popForRequest(test.request, test.currentStatus)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}

			ms.NoError(err, "did not expect any error")

			gotLog := buf.String()

			if test.wantLog == "" {
				ms.Equal("", gotLog, "unexpected logging message")
			} else {
				ms.Contains(gotLog, test.wantLog, "did not get expected logging message")
			}

			var histories RequestHistories
			err = DB.Where("request_id = ?", test.request.ID).All(&histories)
			ms.NoError(err, "unexpected error fetching histories")

			ms.Equal(len(test.want), len(histories), "incorrect number of histories")

			for i := range test.want {
				ms.Equal(test.want[i].Status, histories[i].Status, "incorrect status")
				ms.Equal(test.want[i].ReceiverID, histories[i].ReceiverID, "incorrect receiver id")
			}
		})
	}
}

func (ms *ModelSuite) TestRequestHistory_createForRequest() {
	t := ms.T()
	f := createFixturesForTestRequestHistory_createForRequest(ms)

	tests := []struct {
		name       string
		request    Request
		status     RequestStatus
		providerID int
		wantErr    string
		want       RequestHistories
	}{
		{
			name:       "open to accepted",
			request:    f.Requests[0],
			status:     RequestStatusAccepted,
			providerID: f.Users[1].ID,
			want: RequestHistories{
				{
					Status:     RequestStatusOpen,
					RequestID:  f.Requests[0].ID,
					ReceiverID: nulls.NewInt(f.Requests[0].CreatedByID),
				},
				{
					Status:     RequestStatusAccepted,
					ReceiverID: nulls.NewInt(f.Users[0].ID),
					ProviderID: nulls.NewInt(f.Users[1].ID),
				},
			},
		},
		{
			name:    "null to open",
			request: f.Requests[1],
			status:  RequestStatusOpen,
			want:    RequestHistories{{Status: RequestStatusOpen, ReceiverID: nulls.NewInt(f.Users[0].ID)}},
		},
		{
			name:       "bad provider id",
			request:    f.Requests[1],
			status:     RequestStatusAccepted,
			providerID: 999999,
			wantErr:    `key constraint "request_histories_provider_id_fkey"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.request.Status = test.status

			if test.providerID > 0 {
				test.request.ProviderID = nulls.NewInt(test.providerID)
			}

			var pH RequestHistory

			err := pH.createForRequest(test.request)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}

			ms.NoError(err, "did not expect any error")

			var histories RequestHistories
			err = DB.Where("request_id = ?", test.request.ID).All(&histories)
			ms.NoError(err, "unexpected error fetching histories")

			ms.Equal(len(test.want), len(histories), "incorrect number of histories")

			for i := range test.want {
				ms.Equal(test.want[i].Status, histories[i].Status, "incorrect newStatus")
				ms.Equal(test.want[i].ReceiverID, histories[i].ReceiverID, "incorrect receiver id")

				if test.providerID > 0 {
					ms.Equal(test.want[i].ProviderID, histories[i].ProviderID, "incorrect provider id")
				}
			}
		})
	}
}
