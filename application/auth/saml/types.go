package saml

// SamlUser hold attributes from SAML assertion
type SamlUser struct {
	FirstName string
	LastName  string
	Email     string
	UserID    string
}
