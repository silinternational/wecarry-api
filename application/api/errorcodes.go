package api

const (
	CategoryDatabase  = ErrorCategory("DB")
	CategoryUser      = ErrorCategory("User") // used for errors related to user input, validation, etc.
	CategoryForbidden = ErrorCategory("Forbidden")
	CategoryNotFound  = ErrorCategory("NotFound")
	CategoryInternal  = ErrorCategory("Internal") // used for internal server errors, not related to bad user input
)

const (
	// General

	CreateFailure                   = ErrorKey("ErrorCreateFailure")
	DestroyFailure                  = ErrorKey("ErrorDestroyFailure")
	ErrorGenericInternalServerError = ErrorKey("ErrorGenericInternalServerError")
	FailedToConvertToAPIType        = ErrorKey("ErrorFailedToConvertToAPIType")
	InvalidRequestBody              = ErrorKey("ErrorInvalidRequestBody")
	MustBeAValidUUID                = ErrorKey("ErrorMustBeAValidUUID")
	NoRows                          = ErrorKey("ErrorNoRows")
	NotAuthorized                   = ErrorKey("ErrorNotAuthorized")
	QueryFailure                    = ErrorKey("ErrorQueryFailure")
	SaveFailure                     = ErrorKey("ErrorSaveFailure")
	TransactionNotFound             = ErrorKey("ErrorTransactionNotFound")
	UnknownError                    = ErrorKey("ErrorUnknownError")
	UpdateFailure                   = ErrorKey("ErrorUpdateFailure")
	ValidationError                 = ErrorKey("ErrorValidationError")

	// HTTP codes for customErrorHandler

	ErrorBadRequest           = ErrorKey("ErrorBadRequest")
	ErrorInternalServerError  = ErrorKey("ErrorInternalServerError")
	ErrorMethodNotAllowed     = ErrorKey("ErrorMethodNotAllowed")
	ErrorNotAuthenticated     = ErrorKey("ErrorNotAuthenticated")
	ErrorRouteNotFound        = ErrorKey("ErrorRouteNotFound")
	ErrorUnexpectedHTTPStatus = ErrorKey("ErrorUnexpectedHTTPStatus")
	ErrorUnprocessableEntity  = ErrorKey("ErrorUnprocessableEntity")

	// Authentication

	ErrorAuthEmailMismatch              = ErrorKey("ErrorAuthEmailMismatch")
	ErrorAuthProvidersCallback          = ErrorKey("ErrorAuthProvidersCallback")
	ErrorAuthProvidersLogout            = ErrorKey("ErrorAuthProvidersLogout")
	ErrorCannotFindOrg                  = ErrorKey("ErrorCannotFindOrg")
	ErrorDeletingAccessToken            = ErrorKey("ErrorDeletingAccessToken") // #nosec G101
	ErrorFindingAccessToken             = ErrorKey("ErrorFindingAccessToken")
	ErrorFindingOrgByID                 = ErrorKey("ErrorFindingOrgByID")
	ErrorFindingOrgForAccessToken       = ErrorKey("ErrorFindingOrgForAccessToken")
	ErrorFindingOrgForNewUser           = ErrorKey("ErrorFindingOrgForNewUser")
	ErrorFindingUserByEmail             = ErrorKey("ErrorFindingUserByEmail")
	ErrorFindingUserOrgs                = ErrorKey("ErrorFindingUserOrgs")
	ErrorGettingAuthURL                 = ErrorKey("ErrorGettingAuthURL")
	ErrorGettingSocialAuthUser          = ErrorKey("ErrorGettingSocialAuthUser")
	ErrorInvalidInviteCode              = ErrorKey("ErrorInvalidInviteCode")
	ErrorInvalidInviteType              = ErrorKey("ErrorInvalidInviteType")
	ErrorInvalidSessionInviteObjectUUID = ErrorKey("ErrorInvalidSessionInviteObjectUUID")
	ErrorLoadingAuthProvider            = ErrorKey("ErrorLoadingAuthProvider")
	ErrorMissingAuthEmail               = ErrorKey("ErrorMissingAuthEmail")
	ErrorMissingAuthType                = ErrorKey("ErrorMissingAuthType")
	ErrorMissingClientID                = ErrorKey("ErrorMissingClientID")
	ErrorMissingLogoutToken             = ErrorKey("ErrorMissingLogoutToken")
	ErrorMissingSessionAuthEmail        = ErrorKey("ErrorMissingSessionAuthEmail")
	ErrorMissingSessionClientID         = ErrorKey("ErrorMissingSessionClientID")
	ErrorMissingSessionInviteObjectUUID = ErrorKey("ErrorMissingSessionInviteObjectUUID")
	ErrorMissingSessionSocialAuthType   = ErrorKey("ErrorMissingSessionSocialAuthType")
	ErrorOrglessUserNotAllowed          = ErrorKey("ErrorOrglessUserNotAllowed")
	ErrorWithAuthUser                   = ErrorKey("ErrorWithAuthUser")

	// File

	ErrorReceivingFile           = ErrorKey("ErrorReceivingFile")
	ErrorStoreFileBadContentType = ErrorKey("ErrorStoreFileBadContentType")
	ErrorStoreFileTooLarge       = ErrorKey("ErrorStoreFileTooLarge")
	ErrorUnableToReadFile        = ErrorKey("ErrorUnableToReadFile")
	ErrorUnableToStoreFile       = ErrorKey("ErrorUnableToStoreFile")

	// Message

	MessageBadRequestUUID        = ErrorKey("ErrorMessageBadRequestUUID")
	MessageBadThreadUUID         = ErrorKey("ErrorMessageBadThreadUUID")
	MessageRequestNotVisible     = ErrorKey("ErrorMessageRequestNotVisible")
	MessageThreadNotVisible      = ErrorKey("MessageThreadNotVisible")
	MessageThreadRequestMismatch = ErrorKey("ErrorMessageThreadRequestMismatch")

	// Request

	GetRequests = ErrorKey("GetRequests")

	// Thread

	ThreadsLoadFailure    = ErrorKey("ErrorThreadsLoadFailure")
	ThreadNotFound        = ErrorKey("ErrorThreadNotFound")
	ThreadSetLastViewedAt = ErrorKey("ErrorThreadSetLastViewedAt")

	// User

	UserUpdateError       = ErrorKey("ErrorUserUpdate")
	UserUpdatePhotoError  = ErrorKey("ErrorUserUpdatePhoto")
	UserInvisibleNickname = ErrorKey("ErrorUserInvisibleNickname")
	UserDuplicateNickname = ErrorKey("ErrorUserDuplicateNickname")

	// Watch

	WatchDeleteFailure = ErrorKey("ErrorWatchDeleteFailure")
	WatchesLoadFailure = ErrorKey("ErrorWatchesLoadFailure")
	WatchMissingID     = ErrorKey("ErrorWatchMissingID")
	WatchNotFound      = ErrorKey("ErrorWatchNotFound")
)
