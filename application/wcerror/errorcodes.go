package wcerror

const (
	CategoryDatabase  = ErrorCategory("DB")
	CategoryUser      = ErrorCategory("User") // used for errors related to user input, validation, etc.
	CategoryForbidden = ErrorCategory("Forbidden")
	CategoryNotFound  = ErrorCategory("NotFound")
	CategoryInternal  = ErrorCategory("Internal") // used for internal server errors, not related to bad user input
)

const (
	// =======================  general purpose errors  ==============================
	ErrorGenericInternalServerError = ErrorKey("ErrorGenericInternalServerError")
	FailedToConvertToAPIType        = ErrorKey("ErrorFailedToConvertToAPIType")
	NoRows                          = ErrorKey("ErrorNoRows")
	UnknownError                    = ErrorKey("ErrorUnknownError")

	// ============================  Threads =========================================
	ThreadsLoadFailure = ErrorKey("ErrorThreadsLoadFailure")
)
