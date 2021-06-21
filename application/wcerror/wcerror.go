package wcerror

import (
	"database/sql"
)

type WCerror struct {
	Err      error
	Key      ErrorKey
	Category ErrorCategory
}

type ErrorKey string

func (e ErrorKey) String() string {
	return string(e)
}

type ErrorCategory string

func (e ErrorCategory) String() string {
	return string(e)
}

func (w *WCerror) Error() string {
	if w.Err == nil {
		return ""
	}
	return w.Err.Error()
}

func (w *WCerror) Unwrap() error {
	return w.Err
}

func New(err error, key ErrorKey, category ErrorCategory) *WCerror {
	if err == sql.ErrNoRows {
		key = NoRows
	}
	return &WCerror{
		Err:      err,
		Key:      key,
		Category: category,
	}
}
