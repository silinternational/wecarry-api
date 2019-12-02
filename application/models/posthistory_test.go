package models

import (
	"testing"
)

func (ms *ModelSuite) TestPostHistory_Load() {
	t := ms.T()
	f := createFixturesForTestPostHistory_Load(ms)

	tests := []struct {
		name        string
		postHistory PostHistory
		wantErr     string
		wantEmail   string
	}{
		{
			name:        "open to committed",
			postHistory: f.PostHistories[0],
			wantEmail:   f.Users[0].Email,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.postHistory.Load("Receiver")
			ms.NoError(err, "did not expect any error")

			ms.Equal(test.wantEmail, test.postHistory.Receiver.Email, "incorrect Receiver email")

		})
	}
}
