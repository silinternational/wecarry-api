package domain

import (
	"errors"
	"github.com/gobuffalo/suite"
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

// TestSuite establishes a test suite for domain tests
type TestSuite struct {
	*suite.Model
}

// Test_GqlgenSuite runs the GqlgenSuite test suite
func Test_TestSuite(t *testing.T) {
	model := suite.NewModel()

	gs := &TestSuite{
		Model: model,
	}
	suite.Run(t, gs)
}

func (ts *TestSuite) TestGetFirstStringFromSlice() {
	t := ts.T()

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
			got := GetFirstStringFromSlice(tt.args.s)
			ts.Equal(tt.want, got)
		})
	}
}

func (ts *TestSuite) TestGetBearerTokenFromRequest() {
	t := ts.T()

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
						"Authorization": {"Bearer: 861B1C06-DDB8-494F-8627-3A87B22FFB82"},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBearerTokenFromRequest(tt.args.r)
			ts.Equal(tt.want, got)
		})
	}
}

func (ts *TestSuite) TestGetSubPartKeyValues() {
	t := ts.T()

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
			ts.Equal(tt.want, got)
		})
	}
}

func (ts *TestSuite) TestConvertTimeToStringPtr() {
	t := ts.T()

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
			ts.Equal(test.want, *got)
		})
	}
}

func (ts *TestSuite) TestConvertStringPtrToDate() {
	t := ts.T()

	testTime := time.Date(2019, time.August, 12, 0, 0, 0, 0, time.UTC)
	testStr := testTime.Format("2006-01-02") // not using a const in order to detect code changes
	emptyStr := ""
	badTime := "1"
	type args struct {
		inPtr *string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{{
		name: "nil",
		args: args{nil},
		want: time.Time{},
	}, {
		name: "empty",
		args: args{&emptyStr},
		want: time.Time{},
	}, {
		name: "good",
		args: args{&testStr},
		want: testTime,
	}, {
		name:    "error",
		args:    args{&badTime},
		wantErr: true,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ConvertStringPtrToDate(test.args.inPtr)
			if test.wantErr == false {
				ts.NoError(err)
			}

			ts.Equal(test.want, got)
		})
	}
}

func (ts *TestSuite) TestIsStringInSlice() {
	t := ts.T()

	type testData struct {
		name     string
		needle   string
		haystack []string
		want     bool
	}

	allTestData := []testData{
		{
			name:     "empty haystack",
			needle:   "no",
			haystack: []string{},
			want:     false,
		},
		{
			name:     "not in haystack",
			needle:   "no",
			haystack: []string{"really", "are you sure"},
			want:     false,
		},
		{
			name:     "in one element haystack",
			needle:   "yes",
			haystack: []string{"yes"},
			want:     true,
		},
		{
			name:     "in longer haystack",
			needle:   "yes",
			haystack: []string{"one", "two", "three", "yes"},
			want:     true,
		},
	}

	for i, td := range allTestData {
		t.Run(td.name, func(t *testing.T) {
			got := IsStringInSlice(td.needle, td.haystack)
			ts.Equal(td.want, got, "incorrect value for test %v", i)
		})
	}
}

func (ts *TestSuite) Test_emptyUuidValue() {
	val := uuid.UUID{}
	ts.Equal("00000000-0000-0000-0000-000000000000", val.String(), "incorrect empty uuid value")
}

func (ts *TestSuite) TestEmailDomain() {
	t := ts.T()

	tests := []struct {
		name  string
		email string
		want  string
	}{
		{name: "empty string", email: "", want: ""},
		{name: "domain only", email: "example.org", want: "example.org"},
		{name: "full email", email: "user@example.org", want: "example.org"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := EmailDomain(test.email)
			ts.Equal(test.want, got, "incorrect response from EmailDomain()")
		})
	}
}

func (ts *TestSuite) TestIsOtherThanNoRows() {
	t := ts.T()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "real error", err: errors.New("Real Error"), want: true},
		{name: "no rows error", err: errors.New("sql: no rows in result set"), want: false},
		{name: "nil error", err: nil, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := IsOtherThanNoRows(test.err)
			ts.Equal(test.want, got, "incorrect response")
		})
	}
}

func (ts *TestSuite) TestGetStructFieldTags() {

	type testStruct struct {
		Language string `json:"language"`
		TimeZone string `json:"time_zone"`
	}

	tStruct := testStruct{}

	got, err := GetStructTags("json", tStruct)
	ts.NoError(err)

	gotCount := len(got)
	wantCount := 2
	ts.Equal(wantCount, gotCount, "incorrect number of tags")

	fieldName := "language"
	gotName := got[fieldName]
	ts.Equal(fieldName, gotName, "incorrect tag")

	fieldName = "time_zone"
	gotName = got[fieldName]
	ts.Equal(fieldName, gotName, "incorrect tag")
}

func (ts *TestSuite) TestIsLanguageAllowed() {
	lang := UserPreferenceLanguageSpanish
	got := IsLanguageAllowed(lang)
	ts.True(got, lang+" should be an allowed language")

	lang = "badlanguage"
	got = IsLanguageAllowed(lang)
	ts.False(got, lang+" should not be an allowed language")
}

func (ts *TestSuite) TestIsWeightUnitAllowed() {
	unit := UserPreferenceWeightUnitKGs
	got := IsWeightUnitAllowed(unit)
	ts.True(got, unit+" should be an allowed weight unit")

	unit = "badunit"
	got = IsLanguageAllowed(unit)
	ts.False(got, unit+" should not be an allowed weight unit")
}

func (ts *TestSuite) TestIsTimeZoneAllowed() {
	zone := "America/New_York"
	got := IsTimeZoneAllowed(zone)
	ts.True(got, zone+" should be an allowed time zone")

	zone = "Etc/GMT+2"
	got = IsTimeZoneAllowed(zone)
	ts.True(got, zone+" should be an allowed time zone")

	zone = "badzone"
	got = IsLanguageAllowed(zone)
	ts.False(got, zone+" should not be an allowed language")
}
