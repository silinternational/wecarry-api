package domain

// ********************************************************************
// Don't change the value of these Key entries without making a corresponding change on the UI,
// since these will be converted to human-friendly texts on the UI
// ********************************************************************

// ClientID is required on various requests
// actions.AuthRequest
const ErrorMissingClientID = "ErrorMissingClientID"

// AuthEmail is required on authentication requests
// actions.AuthRequest
const ErrorMissingAuthEmail = "ErrorMissingAuthEmail"

// There was an error during the auth request process when trying to
// find an organization or userorganizations
// actions.AuthRequest
const ErrorFindingOrgUserOrgs = "ErrorFindingOrgUserOrgs"

// An appropriate organization was not found
// for the user making an authentication request
// actions.AuthRequest
const ErrorCannotFindOrg = "ErrorCannotFindOrg"

// actions.AuthRequest
const ErrorSavingAuthRequestSession = "ErrorSavingAuthRequestSession"

// actions.AuthRequest and others
const ErrorLoadingAuthProvider = "ErrorLoadingAuthProvider"

// actions.AuthRequest
const ErrorGettingAuthURL = "ErrorGettingAuthURL"

// actions.AuthCallback
const ErrorMissingSessionAuthEmail = "ErrorMissingSessionAuthEmail"

// actions.AuthCallback
const ErrorMissingSessionClientID = "ErrorMissingSessionClientID"

// actions.AuthCallback
const ErrorMissingSessionOrgID = "ErrorMissingSessionOrgID"

// actions.AuthCallback
const ErrorSavingAuthCallbackSession = "ErrorSavingAuthCallbackSession"

// actions.AuthCallback
const ErrorFindingOrgByID = "ErrorFindingOrgByID"

// actions.AuthCallback
const ErrorAuthProvidersCallback = "ErrorAuthProvidersCallback"

// actions.AuthCallback
const ErrorWithAuthUser = "ErrorWithAuthUser"

// token param is required on a logout request
// actions.AuthDestroy
const ErrorMissingLogoutToken = "ErrorMissingLogoutToken"

// actions.AuthDestroy
const ErrorFindingAccessToken = "ErrorFindingAccessToken"

// actions.AuthDestroy
const ErrorFindingOrgForAccessToken = "ErrorFindingOrgForAccessToken"

// actions.AuthDestroy
const ErrorAuthProvidersLogout = "ErrorAuthProvidersLogout"

// actions.AuthDestroy
const ErrorDeletingAccessToken = "ErrorDeletingAccessToken"

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
