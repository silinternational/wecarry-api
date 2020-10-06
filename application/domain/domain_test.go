package domain

import (
	"errors"
	"github.com/gobuffalo/validate"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

// TestSuite establishes a test suite for domain tests
type TestSuite struct {
	suite.Suite
}

// Test_TestSuite runs the test suite
func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
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

func (ts *TestSuite) Test_emptyUUIDValue() {
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
		{name: "domain only, mixed case", email: "Example.org", want: "example.org"},
		{name: "full email", email: "user@example.org", want: "example.org"},
		{name: "full email, mixed case", email: "User@Example.org", want: "example.org"},
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
	got = IsTimeZoneAllowed(zone)
	ts.False(got, zone+" should not be an time zone")
}

func (ts *TestSuite) TestGetTranslatedSubject() {
	t := ts.T()
	requestTitle := "MyRequest"

	tests := []struct {
		name          string
		language      string
		translationID string
		want          string
	}{
		{
			name:          "delivered",
			translationID: "Email.Subject.Request.FromAcceptedToDelivered",
			want:          `Your ` + Env.AppName + ` request for "` + requestTitle + `" has been delivered!`,
		},
		{
			name:          "delivered in Spanish",
			language:      UserPreferenceLanguageSpanish,
			translationID: "Email.Subject.Request.FromAcceptedToDelivered",
			want:          "Su solicitud se marcó como entregada en " + Env.AppName,
		},
		{
			name:          "from accepted to completed",
			translationID: "Email.Subject.Request.FromAcceptedOrDeliveredToCompleted",
			want:          "Thank you for fulfilling a request on " + Env.AppName,
		},
		{
			name:          "from accepted to open",
			translationID: "Email.Subject.Request.FromAcceptedToOpen",
			want:          `Your ` + Env.AppName + ` offer for "` + requestTitle + `" is no longer needed`,
		},
		{
			name:          "from accepted to removed",
			translationID: "Email.Subject.Request.FromAcceptedToRemoved",
			want:          `Your ` + Env.AppName + ` offer for "` + requestTitle + `" is no longer needed`,
		},
		{
			name:          "from completed to accepted",
			translationID: "Email.Subject.Request.FromCompletedToAcceptedOrDelivered",
			want:          "Request not received on " + Env.AppName + " after all",
		},
		{
			name:          "from delivered to accepted",
			translationID: "Email.Subject.Request.FromDeliveredToAccepted",
			want:          "Request not delivered after all on " + Env.AppName,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			language := UserPreferenceLanguageEnglish
			if test.language != "" {
				language = test.language
			}
			got := GetTranslatedSubject(language, test.translationID,
				map[string]string{"requestTitle": requestTitle})
			ts.Equal(test.want, got, "bad subject translation")
		})
	}
}

func TestTruncate(t *testing.T) {
	type args struct {
		str    string
		suffix string
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string shorter than length, not changed",
			args: args{
				str:    "hello",
				suffix: "...",
				length: 16,
			},
			want: "hello",
		},
		{
			name: "string truncated, empty suffix",
			args: args{
				str:    "hello",
				suffix: "",
				length: 3,
			},
			want: "hel",
		},
		{
			name: "string truncated, with suffix",
			args: args{
				str:    "hello there",
				suffix: "...",
				length: 10,
			},
			want: "hello t...",
		},
		{
			name: "string is length, not truncated",
			args: args{
				str:    "hello there",
				suffix: "...",
				length: 11,
			},
			want: "hello there",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.args.str, tt.args.suffix, tt.args.length); got != tt.want {
				t.Errorf("Truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func (ts *TestSuite) TestEmailFromAddress() {
	nickname := "nickname"

	tests := []struct {
		name string
		arg  *string
		want string
	}{
		{
			name: "name given",
			arg:  &nickname,
			want: "nickname via WeCarry <no_reply@example.com>",
		},
		{
			name: "no name given",
			arg:  nil,
			want: "WeCarry <no_reply@example.com>",
		},
	}
	for _, tt := range tests {
		ts.T().Run(tt.name, func(t *testing.T) {
			if got := EmailFromAddress(tt.arg); got != tt.want {
				t.Errorf("EmailFromAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func (ts *TestSuite) TestRemoveUnwantedChars() {
	tests := []struct {
		name    string
		str     string
		allowed string
		want    string
	}{
		{name: "simple", str: "abc", allowed: "", want: "abc"},
		{name: "none shall pass", str: "!@#$", allowed: "", want: ""},
		{name: "empty str", str: "", allowed: "#", want: ""},
		{name: "Korean", str: string([]rune{0xae40}), want: string([]rune{0xae40})},
		{name: "Flag", str: string([]rune{0x1f1fa, 0x1f1f8}), want: string([]rune{0x1f1fa, 0x1f1f8})},
		{name: "Zero-width joiner", str: string([]rune{0x1f468, 0x200d, 0x1f9b3}),
			want: string([]rune{0x1f468, 0x200d, 0x1f9b3})},
		{name: "Flag", str: string([]rune{0x1f1fa, 0x1f1f8}), want: string([]rune{0x1f1fa, 0x1f1f8})},
		{name: "Chinese", str: string([]rune{0x540d, 0x79f0}), want: ""},
		{name: "Script", str: "<script></script>", allowed: "-_ .,'&", want: "scriptscript"},
		{name: "Spanish", str: "José", allowed: "", want: "José"},
	}
	for _, tt := range tests {
		ts.T().Run(tt.name, func(t *testing.T) {
			ts.Equal(tt.want, RemoveUnwantedChars(tt.str, tt.allowed))
		})
	}
}

func (ts *TestSuite) TestIsSafeRune() {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{name: "Latin Supplement 1", r: 0xc0, want: true},
		{name: "Latin Extended A", r: 0x100, want: true},
		{name: "Zero Width Joiner", r: 0x200d, want: true},
		{name: "Dingbat", r: 0x2700, want: true},
		{name: "Korean", r: 0xac00, want: true},
		{name: "Variation Selector 16", r: 0xfe0f, want: true},
		{name: "Regional Indicator Symbol A", r: 0x1f1e6, want: true},
		{name: "Pictograph", r: 0x1f334, want: true},
		{name: "Emoji", r: 0x1f61b, want: true},
		{name: "Transportation symbol", r: 0x1f6ec, want: true},
		{name: "Supplemental symbol", r: 0x1f91f, want: true},
		{name: "<", r: '<', want: false},
		{name: "CR", r: 0x0d, want: false},
	}
	for _, tt := range tests {
		ts.T().Run(tt.name, func(t *testing.T) {
			ts.Equal(tt.want, isSafeRune(tt.r))
		})
	}
}

func (ts *TestSuite) TestStringIsVisible_IsValid() {
	tests := []struct {
		name    string
		field   string
		message string
		nErrors int
	}{
		{
			name:    "no error",
			field:   "visible string",
			nErrors: 0,
		},
		{
			name:    "empty string, custom message",
			field:   "",
			message: "An error message",
			nErrors: 1,
		},
		{
			name:    "VS16",
			field:   string([]rune{0xfe0f}),
			nErrors: 1,
		},
		{
			name:    "ZWJ",
			field:   string([]rune{0xfe0f}),
			nErrors: 1,
		},
		{
			name:    "ASCII whitespace",
			field:   " \t\n\v\f\r",
			nErrors: 1,
		},
		{
			name:    "visible mixed with invisible",
			field:   string([]rune{0x1f468, 0x200d, 0x1f9b0}),
			nErrors: 0,
		},
	}
	for _, tt := range tests {
		ts.T().Run(tt.name, func(t *testing.T) {
			vErr := validate.NewErrors()

			v := &StringIsVisible{
				Name:    "field_name",
				Field:   tt.field,
				Message: tt.message,
			}
			v.IsValid(vErr)
			ts.Equal(tt.nErrors, len(vErr.Errors))
			if tt.nErrors > 0 {
				if tt.message == "" {
					ts.Equal([]string{v.Name + " must have a visible character."}, vErr.Get(v.Name))
				} else {
					ts.Equal([]string{tt.message}, vErr.Get(v.Name))
				}
			}
		})
	}
}

func (ts *TestSuite) TestUniquifyIntSlice() {
	tests := []struct {
		name string
		ids  []int
		want []int
	}{
		{name: "empty", ids: []int{}, want: []int{}},
		{name: "single", ids: []int{3}, want: []int{3}},
		{name: "all unique", ids: []int{3, 7, 11}, want: []int{3, 7, 11}},
		{name: "all the same", ids: []int{3, 3, 3, 3}, want: []int{3}},
		{name: "some duplicates", ids: []int{3, 4, 5, 4, 3, 3}, want: []int{3, 4, 5}},
	}
	for _, tt := range tests {
		ts.T().Run(tt.name, func(t *testing.T) {
			got := UniquifyIntSlice(tt.ids)
			ts.Equal(tt.want, got, "incorrect int results")
		})
	}
}
