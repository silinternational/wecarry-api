package domain

// ********************************************************************
// Don't change the value of these Key entries without making a corresponding change on the UI,
// since these will be converted to human-friendly texts on the UI
// ********************************************************************

// Unexpected http status code
const ErrorUnexpectedHTTPStatus = "ErrorUnexpectedHTTPStatus"

// 400 http.StatusBadRequest
const ErrorBadRequest = "ErrorBadRequest"

// 401 http.StatusUnauthorized
const ErrorNotAuthenticated = "ErrorNotAuthenticated"

// 404 http.StatusNotFound
const ErrorRouteNotFound = "ErrorRouteNotFound"

// 405 http.StatusMethodNotAllowed
const ErrorMethodNotAllowed = "ErrorMethodNotAllowed"

// 422 http.StatusUnprocessableEntity
const ErrorUnprocessableEntity = "ErrorUnprocessableEntity"

// 500 http.StatusInternalServerError
const ErrorInternalServerError = "ErrorInternalServerError"

// ClientID is required on various requests
// actions.authRequest
const ErrorMissingClientID = "ErrorMissingClientID"

// AuthEmail is required on authentication requests
// actions.authRequest
const ErrorMissingAuthEmail = "ErrorMissingAuthEmail"

// AuthType is required on social authentication select calls
// actions.authSelect
const ErrorMissingAuthType = "ErrorMissingAuthType"

// There was an error during the auth request process when trying to
// find an organization for a user with no UserOrganizations
// actions.AuthRequest
const ErrorFindingOrgForNewUser = "ErrorFindingOrgForNewUser"

// There was an error during the auth request process when trying to
// find userorganizations
// actions.authRequest and actions.meetingAuthRequest
const ErrorFindingUserOrgs = "ErrorFindingUserOrgs"

// actions.finishAuthRequestForSocialUser
const ErrorFindingUserByEmail = "ErrorFindingUserByEmail"

// No Organization was found for the authEmail
// actions.authRequest
const ErrorOrglessUserNotAllowed = "ErrorOrglessUserNotAllowed"

// An appropriate organization was not found
// for the user making an authentication request
// actions.inviteAuthRequest
const ErrorCannotFindOrg = "ErrorCannotFindOrg"

// actions.inviteAuthRequest
const ErrorInvalidInviteType = "ErrorInvalidInviteType"

// actions.getAuthInviteResponse
const ErrorInvalidInviteCode = "ErrorInvalidInviteCode"

// actions.meetingAuthRequest
const ErrorInvalidSessionInviteObjectUUID = "ErrorInvalidSessionInviteObjectUUID"

// actions - various places
const ErrorLoadingAuthProvider = "ErrorLoadingAuthProvider"

// actions.getOrgBasedAuthOption, actions.authSelect, and actions.finishAuthRequestForSocialUser
const ErrorGettingAuthURL = "ErrorGettingAuthURL"

// actions.meetingAuthRequest
const ErrorMissingSessionInviteObjectUUID = "ErrorMissingSessionInviteObjectUUID"

// actions.authCallback
const ErrorMissingSessionAuthEmail = "ErrorMissingSessionAuthEmail"

// actions.authCallback
const ErrorMissingSessionClientID = "ErrorMissingSessionClientID"

// actions.socialLoginBasedAuthCallback
const ErrorMissingSessionSocialAuthType = "ErrorMissingSessionSocialAuthType"

// actions.orgBasedAuthCallback
const ErrorFindingOrgByID = "ErrorFindingOrgByID"

// actions several locations
const ErrorAuthProvidersCallback = "ErrorAuthProvidersCallback"

// actions.orgBasedAuthCallback
const ErrorAuthEmailMismatch = "ErrorAuthEmailMismatch"

// actions.socialLoginNonInviteBasedAuthCallback
const ErrorGettingSocialAuthUser = "ErrorGettingSocialAuthUser"

// actions.orgBasedAuthCallback and actions.socialLoginBasedAuthCallback
const ErrorWithAuthUser = "ErrorWithAuthUser"

// token param is required on a logout request
// actions.authDestroy
const ErrorMissingLogoutToken = "ErrorMissingLogoutToken"

// actions.authDestroy
const ErrorFindingAccessToken = "ErrorFindingAccessToken"

// actions.authDestroy
const ErrorFindingOrgForAccessToken = "ErrorFindingOrgForAccessToken"

// actions.authDestroy
const ErrorAuthProvidersLogout = "ErrorAuthProvidersLogout"

// actions.authDestroy
const ErrorDeletingAccessToken = "ErrorDeletingAccessToken" // #nosec G101

// actions.UploadHandler
const ErrorReceivingFile = "ErrorReceivingFile"

// actions.UploadHandler
const ErrorUnableToReadFile = "ErrorUnableToReadFile"

// actions.UploadHandler
const ErrorUnableToStoreFile = "ErrorUnableToStoreFile"

// models.Store
const ErrorStoreFileTooLarge = "ErrorStoreFileTooLarge"

// models.Store
const ErrorStoreFileBadContentType = "ErrorStoreFileBadContentType"
