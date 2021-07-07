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

	ErrorCreateFailure            = ErrorKey("ErrorCreateFailure")
	ErrorDestroyFailure           = ErrorKey("ErrorDestroyFailure")
	ErrorGenericInternalServer    = ErrorKey("ErrorGenericInternalServer")
	ErrorFailedToConvertToAPIType = ErrorKey("ErrorFailedToConvertToAPIType")
	ErrorInvalidRequestBody       = ErrorKey("ErrorInvalidRequestBody")
	ErrorMustBeAValidUUID         = ErrorKey("ErrorMustBeAValidUUID")
	ErrorNoRows                   = ErrorKey("ErrorNoRows")
	ErrorNotAuthorized            = ErrorKey("ErrorNotAuthorized")
	ErrorQueryFailure             = ErrorKey("ErrorQueryFailure")
	ErrorSaveFailure              = ErrorKey("ErrorSaveFailure")
	ErrorTransactionNotFound      = ErrorKey("ErrorTransactionNotFound")
	ErrorUnknown                  = ErrorKey("ErrorUnknown")
	ErrorUpdateFailure            = ErrorKey("ErrorUpdateFailure")
	ErrorValidation               = ErrorKey("ErrorValidation")

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

	// Location

	ErrorLocationCreateFailure = ErrorKey("ErrorLocationCreateFailure")

	// Meeting

	ErrorGetMeetings = ErrorKey("ErrorGetMeetings")

	// Message

	ErrorMessageBadRequestUUID        = ErrorKey("ErrorMessageBadRequestUUID")
	ErrorMessageBadThreadUUID         = ErrorKey("ErrorMessageBadThreadUUID")
	ErrorMessageRequestNotVisible     = ErrorKey("ErrorMessageRequestNotVisible")
	ErrorMessageThreadNotVisible      = ErrorKey("ErrorMessageThreadNotVisible")
	ErrorMessageThreadRequestMismatch = ErrorKey("ErrorMessageThreadRequestMismatch")

	// Request

	ErrorGetRequests              = ErrorKey("ErrorGetRequests")
	ErrorGetRequest               = ErrorKey("ErrorGetRequest")
	ErrorGetRequestUserNotAllowed = ErrorKey("ErrorGetRequestUserNotAllowed")

	// Thread

	ErrorThreadsLoadFailure    = ErrorKey("ErrorThreadsLoadFailure")
	ErrorThreadNotFound        = ErrorKey("ErrorThreadNotFound")
	ErrorThreadSetLastViewedAt = ErrorKey("ErrorThreadSetLastViewedAt")

	// User

	ErrorUserUpdate            = ErrorKey("ErrorUserUpdate")
	ErrorUserUpdatePhoto       = ErrorKey("ErrorUserUpdatePhoto")
	ErrorUserInvisibleNickname = ErrorKey("ErrorUserInvisibleNickname")
	ErrorUserDuplicateNickname = ErrorKey("ErrorUserDuplicateNickname")

	// Watch

	ErrorWatchCreateFailure       = ErrorKey("ErrorWatchCreateFailure")
	ErrorWatchDeleteFailure       = ErrorKey("ErrorWatchDeleteFailure")
	ErrorWatchInputEmpty          = ErrorKey("ErrorWatchInputEmpty")
	ErrorWatchInputMeetingFailure = ErrorKey("ErrorWatchInputMeetingFailure")
	ErrorWatchesLoadFailure       = ErrorKey("ErrorWatchesLoadFailure")
	ErrorWatchMissingID           = ErrorKey("ErrorWatchMissingID")
	ErrorWatchNotFound            = ErrorKey("ErrorWatchNotFound")
)
