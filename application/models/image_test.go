package models

import (
	"testing"

	"github.com/gobuffalo/validate"
	"github.com/silinternational/handcarry-api/domain"
)

func (ms *ModelSuite) TestImage_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		image    Image
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			image: Image{
				UUID:   domain.GetUuid(),
				PostID: 1,
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			image: Image{
				PostID: 1,
			},
			wantErr:  true,
			errField: "UUID",
		},
		{
			name: "missing post_id",
			image: Image{
				UUID: domain.GetUuid(),
			},
			wantErr:  true,
			errField: "post_id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.image.Validate(DB)
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
