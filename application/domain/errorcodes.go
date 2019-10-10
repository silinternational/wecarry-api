package domain

// ClientID is required on various requests
// actions.AuthRequest
const MissingClientID = "MissingClientID"

// AuthEmail is required on authentication requests
// actions.AuthRequest
const MissingAuthEmail = "MissingAuthEmail"

// There was an error during the auth request process when trying to
// find an organization or userorganizations
// actions.AuthRequest
const ErrorFindingOrgUserOrgs = "ErrorFindingOrgUserOrgs"

// An appropriate organization was not found
// for the user making an authentication request
// actions.AuthRequest
const CannotFindOrg = "CannotFindOrg"

// actions.AuthRequest
const ErrorSavingAuthRequestSession = "ErrorSavingAuthRequestSession"

// actions.AuthRequest and others
const ErrorLoadingAuthProvider = "ErrorLoadingAuthProvider"

// actions.AuthRequest
const ErrorGettingAuthURL = "ErrorGettingAuthURL"

// actions.AuthCallback
const MissingSessionAuthEmail = "MissingSessionAuthEmail"

// actions.AuthCallback
const MissingSessionClientID = "MissingSessionClientID"

// actions.AuthCallback
const MissingSessionOrgID = "MissingSessionOrgID"

// actions.AuthCallback
const ErrorSavingAuthCallbackSession = "ErrorSavingAuthCallbackSession"

// actions.AuthCallback
const ErrorFindingOrg = "ErrorFindingOrg"

// actions.AuthCallback
const ErrorAuthProvidersCallback = "ErrorAuthProvidersCallback"

// actions.AuthCallback
const ErrorWithAuthUser = "ErrorWithAuthUser"

// token param is required on a logout request
// actions.AuthDestroy
const MissingLogoutToken = "MissingLogoutToken"

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
const UnableToReadFile = "UnableToReadFile"

// actions.UploadHandler
const UnableToStoreFile = "UnableToStoreFile"
