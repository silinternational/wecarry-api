package models

import "testing"

func (ms *ModelSuite) TestPostSize_isLargerOrSame() {
	tests := []struct {
		name string
		p    PostSize
		p2   PostSize
		want bool
	}{
		{
			name: "same",
			p:    PostSizeTiny,
			p2:   PostSizeTiny,
			want: true,
		},
		{
			name: "larger",
			p:    PostSizeSmall,
			p2:   PostSizeTiny,
			want: true,
		},
		{
			name: "smaller",
			p:    PostSizeTiny,
			p2:   PostSizeSmall,
			want: false,
		},
		{
			name: "p empty",
			p:    PostSize(""),
			p2:   PostSizeXlarge,
			want: true,
		},
		{
			name: "p2 empty",
			p:    PostSizeXlarge,
			p2:   PostSize(""),
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
