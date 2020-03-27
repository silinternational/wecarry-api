package models

import "testing"

func (ms *ModelSuite) TestPostSize_isLargerOrSame() {
	tests := []struct {
		name string
		p    RequestSize
		p2   RequestSize
		want bool
	}{
		{
			name: "same",
			p:    RequestSizeTiny,
			p2:   RequestSizeTiny,
			want: true,
		},
		{
			name: "larger",
			p:    RequestSizeSmall,
			p2:   RequestSizeTiny,
			want: true,
		},
		{
			name: "smaller",
			p:    RequestSizeTiny,
			p2:   RequestSizeSmall,
			want: false,
		},
		{
			name: "p empty",
			p:    RequestSize(""),
			p2:   RequestSizeXlarge,
			want: true,
		},
		{
			name: "p2 empty",
			p:    RequestSizeXlarge,
			p2:   RequestSize(""),
			want: false,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			if got := tt.p.isLargerOrSame(tt.p2); got != tt.want {
				t.Errorf("isLargerOrSame() = %v, want = %v", got, tt.want)
			}
		})
	}
}
