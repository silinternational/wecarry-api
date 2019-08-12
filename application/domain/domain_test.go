package domain

import (
	"bytes"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestGetRequestData(t *testing.T) {
	// test GET request
	req := http.Request{
		Method: "GET",
		URL:    &url.URL{RawQuery: "param=val"},
	}
	data, err := GetRequestData(&req)
	if err != nil {
		t.Errorf("GetRequestData() error: %v", err)
	}
	if v, ok := data["param"]; ok {
		if len(v) != 1 || v[0] != "val" {
			t.Errorf("Invalid data: %v", v)
		}
	} else {
		t.Errorf("Missing parameter")
	}

	// test POST request
	body := []byte("param=val")
	postRequest, err := http.NewRequest("POST", "http://www.google.com", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("NewRequest() error: %v", err)
	}
	postRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	data, err = GetRequestData(postRequest)
	if err != nil {
		t.Errorf("GetRequestData() error: %v", err)
	}
	if v, ok := data["param"]; ok {
		if len(v) != 1 || v[0] != "val" {
			t.Errorf("Invalid data: %v", v)
		}
	} else {
		t.Errorf("Missing parameter (data: %v) (body: %v)", data, string(body))
	}
}

func TestGetFirstStringFromSlice(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil",
			args: args{
				s: nil,
			},
			want: "",
		},
		{
			name: "empty slice",
			args: args{
				s: []string{},
			},
			want: "",
		},
		{
			name: "single string in slice",
			args: args{
				s: []string{"alpha"},
			},
			want: "alpha",
		},
		{
			name: "two strings in slice",
			args: args{
				s: []string{"alpha", "beta"},
			},
			want: "alpha",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFirstStringFromSlice(tt.args.s); got != tt.want {
				t.Errorf("GetFirstStringFromSlice() = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

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
			name: "also valid, not case-sensitive",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"bearer def456"},
					},
				},
			},
			want: "def456",
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
						"Authorization": {"Bearer 861B1C06-DDB8-494F-8627-3A87B22FFB82"},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBearerTokenFromRequest(tt.args.r); got != tt.want {
				t.Errorf("GetBearerTokenFromRequest() = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

func TestGetSubPartKeyValues(t *testing.T) {
	type args struct {
		inString, outerDelimiter, innerDelimiter string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "empty string",
			args: args{
				inString:       "",
				outerDelimiter: "!",
				innerDelimiter: "*",
			},
			want: map[string]string{},
		},
		{
			name: "one pair",
			args: args{
				inString:       "param^value",
				outerDelimiter: "#",
				innerDelimiter: "^",
			},
			want: map[string]string{
				"param": "value",
			},
		},
		{
			name: "two pairs",
			args: args{
				inString:       "param1(value1@param2(value2",
				outerDelimiter: "@",
				innerDelimiter: "(",
			},
			want: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
		},
		{
			name: "no inner delimiter",
			args: args{
				inString:       "param-value",
				outerDelimiter: "-",
				innerDelimiter: "=",
			},
			want: map[string]string{},
		},
		{
			name: "extra inner delimiter",
			args: args{
				inString:       "param=value=extra",
				outerDelimiter: "-",
				innerDelimiter: "=",
			},
			want: map[string]string{},
		},
		{
			name: "empty value",
			args: args{
				inString:       "param=value-empty=",
				outerDelimiter: "-",
				innerDelimiter: "=",
			},
			want: map[string]string{
				"param": "value",
				"empty": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSubPartKeyValues(tt.args.inString, tt.args.outerDelimiter, tt.args.innerDelimiter)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSubPartKeyValues() = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

func TestConvertTimeToStringPtr(t *testing.T) {
	now := time.Now()
	type args struct {
		inTime time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default",
			args: args{
				inTime: time.Time{},
			},
			want: "0001-01-01T00:00:00Z",
		},
		{
			name: "now",
			args: args{
				inTime: now,
			},
			want: now.Format(time.RFC3339),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ConvertTimeToStringPtr(test.args.inTime)
			if *got != test.want {
				t.Errorf("ConvertTimeToStringPtr() = \"%v\", want \"%v\"", *got, test.want)
			}
		})
	}
}

func TestIsStringInSlice(t *testing.T) {
	type TestData struct {
		Needle   string
		Haystack []string
		Expected bool
	}

	allTestData := []TestData{
		{
			Needle:   "no",
			Haystack: []string{},
			Expected: false,
		},
		{
			Needle:   "no",
			Haystack: []string{"really", "are you sure"},
			Expected: false,
		},
		{
			Needle:   "yes",
			Haystack: []string{"yes"},
			Expected: true,
		},
		{
			Needle:   "yes",
			Haystack: []string{"one", "two", "three", "yes"},
			Expected: true,
		},
	}

	for i, td := range allTestData {
		results := IsStringInSlice(td.Needle, td.Haystack)
		expected := td.Expected

		if results != expected {
			t.Errorf("Bad results for test set i = %v. Expected %v, but got %v", i, expected, results)
			return
		}
	}
}
