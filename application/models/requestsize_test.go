package models

import "testing"

func (ms *ModelSuite) TestRequestSize_isLargerOrSame() {
	tests := []struct {
		name string
		r    RequestSize
		r2   RequestSize
		want bool
	}{
		{
			name: "same",
			r:    RequestSizeTiny,
			r2:   RequestSizeTiny,
			want: true,
		},
		{
			name: "larger",
			r:    RequestSizeSmall,
			r2:   RequestSizeTiny,
			want: true,
		},
		{
			name: "smaller",
			r:    RequestSizeTiny,
			r2:   RequestSizeSmall,
			want: false,
		},
		{
			name: "p empty",
			r:    RequestSize(""),
			r2:   RequestSizeXlarge,
			want: true,
		},
		{
			name: "r2 empty",
			r:    RequestSizeXlarge,
			r2:   RequestSize(""),
			want: false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			if got := tt.r.isLargerOrSame(tt.r2); got != tt.want {
				t.Errorf("isLargerOrSame() = %v, want = %v", got, tt.want)
			}
		})
	}
}
