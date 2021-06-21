package api

import (
	"database/sql"
	"net/http"
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
