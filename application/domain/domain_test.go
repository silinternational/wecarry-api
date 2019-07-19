package domain

import (
	"net/http"
	"testing"
)

func TestGetBearerTokenFromRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"Bearer abc123"},
					},
				},
			},
			want: "abc123",
		},
		{
			name: "missing authorization header",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Other": {"Bearer abc123"},
					},
				},
			},
			want: "",
		},
		{
			name: "valid, but more complicated",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"Bearer 861B1C06-DDB8-494F-8627-3A87B22FFB82"},
					},
				},
			},
			want: "861B1C06-DDB8-494F-8627-3A87B22FFB82",
		},
		{
			name: "invalid format, missing bearer",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"861B1C06-DDB8-494F-8627-3A87B22FFB82"},
					},
				},
			},
			want: "",
		},
		{
			name: "invalid format, has colon",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"Bearer: 861B1C06-DDB8-494F-8627-3A87B22FFB82"},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBearerTokenFromRequest(tt.args.r); got != tt.want {
				t.Errorf("GetBearerTokenFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
