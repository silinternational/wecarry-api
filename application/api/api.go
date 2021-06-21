package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/domain"
)

type ErrorKey string

func (e ErrorKey) String() string {
	return string(e)
}

type ErrorCategory string

func (e ErrorCategory) String() string {
	return string(e)
}

type AppError struct {
	Code int `json:"code"`

	// Don't change the value of these Key entries without making a corresponding change on the UI,
	// since these will be converted to human-friendly texts for presentation to the user
	Key ErrorKey `json:"key"`

	HttpStatus int `json:"status"`

	// detailed error message for debugging
	Err error `json:"debug_msg,omitempty"`

	Category ErrorCategory `json:"-"`

	Message string `json:"message"`

	// Extra data providing detail about the error condition, only provided in development mode
	Extras map[string]interface{} `json:"extras,omitempty"`

	// URL to redirect, if HttpStatus is in 300-series
	RedirectURL string `json:"-"`
}

func (a *AppError) Error() string {
	if a.Err == nil {
		return ""
	}
	return a.Err.Error()
}

func (a *AppError) Unwrap() error {
	return a.Err
}

func NewAppError(err error, key ErrorKey, category ErrorCategory) *AppError {
	if err == sql.ErrNoRows {
		key = NoRows
	}
	a := AppError{
		Err:      err,
		Key:      key,
		Category: category,
	}
	a.SetHttpStatusFromCategory()
	return &a
}

func (a *AppError) SetHttpStatusFromCategory() {
	switch a.Category {
	case CategoryInternal, CategoryDatabase:
		a.HttpStatus = http.StatusInternalServerError
	case CategoryForbidden, CategoryNotFound:
		a.HttpStatus = http.StatusNotFound
	default:
		a.HttpStatus = http.StatusBadRequest
	}
}

func (a *AppError) LoadTranslatedMessage(c buffalo.Context) {
	if a.HttpStatus == http.StatusInternalServerError {
		errKey := "Error." + ErrorGenericInternalServerError.String()
		a.Message = domain.T.Translate(c, errKey, a.Extras)
		return
	}

	msgID := fmt.Sprintf("Error.%s", a.Key)
	a.Message = domain.T.Translate(c, msgID, a.Extras)
	if a.Message == msgID {
		a.Message = keyToReadableString(a.Key.String())
	}
}

// keyToReadableString takes a key like ErrorSomethingSomethingOther and returns Error something something other
func keyToReadableString(key string) string {
	re := regexp.MustCompile(`[A-Z][^A-Z]*`)
	words := re.FindAllString(key, -1)

	if len(words) == 0 {
		return key
	}

	if len(words) == 1 {
		return words[0]
	}

	for i := 1; i < len(words); i++ {
		words[i] = strings.ToLower(words[i])
	}

	return strings.Join(words, " ")
}

// ConvertToOtherType uses json marshal/unmarshal to convert one type to another.
// Output parameter should be a pointer to the receiving struct
func ConvertToOtherType(input, output interface{}) error {
	str, err := json.Marshal(input)
	if err != nil {
		return NewAppError(
			fmt.Errorf("failed to convert to apitype. marshal error: %s", err.Error()),
			FailedToConvertToAPIType,
			CategoryInternal,
		)
	}
	if err := json.Unmarshal(str, output); err != nil {
		return NewAppError(
			fmt.Errorf("failed to convert to apitype. unmarshal error: %s", err.Error()),
			FailedToConvertToAPIType,
			CategoryInternal,
		)
	}

	return nil
}
