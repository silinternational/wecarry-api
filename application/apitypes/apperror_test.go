package apitypes

import (
	"testing"

	"github.com/silinternational/wecarry-api/wcerror"
)

func Test_keyToReadableString(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: wcerror.ErrorGenericInternalServerError.String(),
			key:  wcerror.ErrorGenericInternalServerError.String(),
			want: "Error generic internal server error",
		},
		{
			name: wcerror.NoRows.String(),
			key:  wcerror.NoRows.String(),
			want: "Error no rows",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyToReadableString(tt.key); got != tt.want {
				t.Errorf("keyToReadableString() = %v, want %v", got, tt.want)
			}
		})
	}
}
