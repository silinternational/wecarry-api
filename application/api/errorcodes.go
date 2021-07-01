package api

const (
	CategoryDatabase  = ErrorCategory("DB")
	CategoryUser      = ErrorCategory("User") // used for errors related to user input, validation, etc.
	CategoryForbidden = ErrorCategory("Forbidden")
	CategoryNotFound  = ErrorCategory("NotFound")
	CategoryInternal  = ErrorCategory("Internal") // used for internal server errors, not related to bad user input
)

const (
	// =======================  general purpose errors  ==============================

	CreateFailure                   = ErrorKey("ErrorCreateFailure")
	DestroyFailure                  = ErrorKey("ErrorDestroyFailure")
	InvalidRequestBody              = ErrorKey("ErrorInvalidRequestBody")
	MustBeAValidUUID                = ErrorKey("ErrorMustBeAValidUUID")
	NoRows                          = ErrorKey("ErrorNoRows")
	NotAuthorized                   = ErrorKey("ErrorNotAuthorized")
	SaveFailure                     = ErrorKey("ErrorSaveFailure")
	QueryFailure                    = ErrorKey("ErrorQueryFailure")
	TransactionNotFound             = ErrorKey("ErrorTransactionNotFound")
	UnknownError                    = ErrorKey("ErrorUnknownError")
	UpdateFailure                   = ErrorKey("ErrorUpdateFailure")
	ValidationError                 = ErrorKey("ErrorValidationError")
	ConfigurationError              = ErrorKey("ErrorConfigurationError")
	FailedToConvertToAPIType        = ErrorKey("ErrorFailedToConvertToAPIType")
	ErrorGenericInternalServerError = ErrorKey("ErrorGenericInternalServerError")

	// ============================  Locations =========================================
	LocationCreateFailure = ErrorKey("ErrorLocationCreateFailure")

	// ============================  Messages =======================================
	MessageBadRequestUUID        = ErrorKey("ErrorMessageBadRequestUUID")
	MessageBadThreadUUID         = ErrorKey("ErrorMessageBadThreadUUID")
	MessageRequestNotVisible     = ErrorKey("ErrorMessageRequestNotVisible")
	MessageThreadRequestMismatch = ErrorKey("ErrorMessageThreadRequestMismatch")
	MessageThreadNotVisible      = ErrorKey("MessageThreadNotVisible")

	// ============================  Threads =========================================
	ThreadsLoadFailure    = ErrorKey("ErrorThreadsLoadFailure")
	ThreadNotFound        = ErrorKey("ErrorThreadNotFound")
	ThreadSetLastViewedAt = ErrorKey("ErrorThreadSetLastViewedAt")

	// ============================  Users =========================================
	UserUpdateError       = ErrorKey("ErrorUserUpdate")
	UserUpdatePhotoError  = ErrorKey("ErrorUserUpdatePhoto")
	UserInvisibleNickname = ErrorKey("ErrorUserInvisibleNickname")
	UserDuplicateNickname = ErrorKey("ErrorUserDuplicateNickname")

	// ============================  Watches =========================================
	WatchInputEmpty          = ErrorKey("ErrorWatchInputEmpty")
	WatchInputMeetingFailure = ErrorKey("ErrorWatchInputMeetingFailure")
	WatchCreateFailure       = ErrorKey("ErrorWatchCreateFailure")
	WatchDeleteFailure       = ErrorKey("ErrorWatchDeleteFailure")
	WatchesLoadFailure       = ErrorKey("ErrorWatchesLoadFailure")
	WatchMissingID           = ErrorKey("ErrorWatchMissingID")
	WatchNotFound            = ErrorKey("ErrorWatchNotFound")
)

// ********************************************************************
// Don't change the value of these Key entries without making a corresponding change on the UI,
// since these will be converted to human-friendly texts on the UI
// ********************************************************************

// Unexpected http status code
const ErrorUnexpectedHTTPStatus = ErrorKey("ErrorUnexpectedHTTPStatus")

// 400 http.StatusBadRequest
const ErrorBadRequest = ErrorKey("ErrorBadRequest")

// 401 http.StatusUnauthorized
const ErrorNotAuthenticated = ErrorKey("ErrorNotAuthenticated")

// 404 http.StatusNotFound
const ErrorRouteNotFound = ErrorKey("ErrorRouteNotFound")

// 405 http.StatusMethodNotAllowed
const ErrorMethodNotAllowed = ErrorKey("ErrorMethodNotAllowed")

// 422 http.StatusUnprocessableEntity
const ErrorUnprocessableEntity = ErrorKey("ErrorUnprocessableEntity")

// 500 http.StatusInternalServerError
const ErrorInternalServerError = ErrorKey("ErrorInternalServerError")

// ClientID is required on various requests
// actions.authRequest
const ErrorMissingClientID = ErrorKey("ErrorMissingClientID")

// AuthEmail is required on authentication requests
// actions.authRequest
const ErrorMissingAuthEmail = ErrorKey("ErrorMissingAuthEmail")

// AuthType is required on social authentication select calls
// actions.authSelect
const ErrorMissingAuthType = ErrorKey("ErrorMissingAuthType")

// There was an error during the auth request process when trying to
// find an organization for a user with no UserOrganizations
// actions.AuthRequest
const ErrorFindingOrgForNewUser = ErrorKey("ErrorFindingOrgForNewUser")

// There was an error during the auth request process when trying to
// find userorganizations
// actions.authRequest and actions.meetingAuthRequest
const ErrorFindingUserOrgs = ErrorKey("ErrorFindingUserOrgs")

// actions.finishAuthRequestForSocialUser
const ErrorFindingUserByEmail = ErrorKey("ErrorFindingUserByEmail")

// No Organization was found for the authEmail
// actions.authRequest
const ErrorOrglessUserNotAllowed = ErrorKey("ErrorOrglessUserNotAllowed")

// An appropriate organization was not found
// for the user making an authentication request
// actions.inviteAuthRequest
const ErrorCannotFindOrg = ErrorKey("ErrorCannotFindOrg")

// actions.inviteAuthRequest
const ErrorInvalidInviteType = ErrorKey("ErrorInvalidInviteType")

// actions.getAuthInviteResponse
const ErrorInvalidInviteCode = ErrorKey("ErrorInvalidInviteCode")

// actions.meetingAuthRequest
const ErrorInvalidSessionInviteObjectUUID = ErrorKey("ErrorInvalidSessionInviteObjectUUID")

// actions - various places
const ErrorLoadingAuthProvider = ErrorKey("ErrorLoadingAuthProvider")

// actions.getOrgBasedAuthOption, actions.authSelect, and actions.finishAuthRequestForSocialUser
const ErrorGettingAuthURL = ErrorKey("ErrorGettingAuthURL")

// actions.meetingAuthRequest
const ErrorMissingSessionInviteObjectUUID = ErrorKey("ErrorMissingSessionInviteObjectUUID")

// actions.authCallback
const ErrorMissingSessionAuthEmail = ErrorKey("ErrorMissingSessionAuthEmail")

// actions.authCallback
const ErrorMissingSessionClientID = ErrorKey("ErrorMissingSessionClientID")

// actions.socialLoginBasedAuthCallback
const ErrorMissingSessionSocialAuthType = ErrorKey("ErrorMissingSessionSocialAuthType")

// actions.orgBasedAuthCallback
const ErrorFindingOrgByID = ErrorKey("ErrorFindingOrgByID")

// actions several locations
const ErrorAuthProvidersCallback = ErrorKey("ErrorAuthProvidersCallback")

// actions.orgBasedAuthCallback
const ErrorAuthEmailMismatch = ErrorKey("ErrorAuthEmailMismatch")

// actions.socialLoginNonInviteBasedAuthCallback
const ErrorGettingSocialAuthUser = ErrorKey("ErrorGettingSocialAuthUser")

// actions.orgBasedAuthCallback and actions.socialLoginBasedAuthCallback
const ErrorWithAuthUser = ErrorKey("ErrorWithAuthUser")

// token param is required on a logout request
// actions.authDestroy
const ErrorMissingLogoutToken = ErrorKey("ErrorMissingLogoutToken")

// actions.authDestroy
const ErrorFindingAccessToken = ErrorKey("ErrorFindingAccessToken")

// actions.authDestroy
const ErrorFindingOrgForAccessToken = ErrorKey("ErrorFindingOrgForAccessToken")

// actions.authDestroy
const ErrorAuthProvidersLogout = ErrorKey("ErrorAuthProvidersLogout")

// actions.authDestroy
const ErrorDeletingAccessToken = ErrorKey("ErrorDeletingAccessToken") // #nosec G101

// actions.UploadHandler
const ErrorReceivingFile = ErrorKey("ErrorReceivingFile")

// actions.UploadHandler
const ErrorUnableToReadFile = ErrorKey("ErrorUnableToReadFile")

// actions.UploadHandler
const ErrorUnableToStoreFile = ErrorKey("ErrorUnableToStoreFile")

// models.Store
const ErrorStoreFileTooLarge = ErrorKey("ErrorStoreFileTooLarge")

// models.Store
const ErrorStoreFileBadContentType = ErrorKey("ErrorStoreFileBadContentType")
