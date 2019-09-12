package models

import (
	"testing"

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
