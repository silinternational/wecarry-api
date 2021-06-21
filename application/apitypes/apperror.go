package apitypes

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/wcerror"
)

// swagger:model
type AppError struct {
	HttpStatus int `json:"-"`

	// Key that maps errors to human-friendly text on the UI
	Key wcerror.ErrorKey `json:"key"`

	// Error message to be presented to end user.
	// Set automatically based on Key and localized version of error message
	Message string `json:"message"`

	// Message providing detail about the error condition, only provided in development mode
	DebugMsg string `json:"debug_msg,omitempty"`

	// Extra data providing detail about the error condition, only provided in development mode
	Extras map[string]interface{} `json:"extras,omitempty"`

	// URL to redirect, if HttpStatus is in 300-series
	RedirectURL string `json:"-"`
}

func (a *AppError) Error() string {
	return a.DebugMsg
}

func (a *AppError) LoadTranslatedMessage(c buffalo.Context) {
	if a.HttpStatus == http.StatusInternalServerError {
		errKey := "Error." + wcerror.ErrorGenericInternalServerError.String()
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
