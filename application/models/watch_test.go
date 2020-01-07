package models

import (
	"testing"

	"github.com/gobuffalo/validate"

	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestWatch_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		watch    Watch
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			watch: Watch{
				UUID:   domain.GetUUID(),
				UserID: 1,
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			watch: Watch{
				UserID: 1,
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "missing user_id",
			watch: Watch{
				UUID: domain.GetUUID(),
			},
			wantErr:  true,
			errField: "user_id",
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
