package saml2provider

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/markbates/goth"
	"golang.org/x/oauth2"

	"fmt"
	"io/ioutil"

	"crypto/x509"

	"encoding/base64"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
)

const IDPMetadataFile = "./samlmetadata/idp-metadata.xml"
const CallbackURL = "http://handcarry.local:3000/auth/saml2/callback/"
const SPIssuer = "http://handcarry.local:3000"
const SPAudienceURI = SPIssuer
const SPReturnTo = "handcarry.local:3000/"

// Provider needs to be implemented for each 3rd party authentication provider
// e.g. Facebook, Twitter, etc...
//type Provider interface {
//	Name() string
//	SetName(name string)
//	BeginAuth(state string) (Session, error)
//	UnmarshalSession(string) (Session, error)
//	FetchUser(Session) (User, error)
//	Debug(bool)
//	RefreshToken(refreshToken string) (*oauth2.Token, error) //Get new access token based on the refresh token
//	RefreshTokenAvailable() bool                             //Refresh token is provided by auth provider or not
//}

type Provider struct {
	providerName string
	IDPURL       string
	CallbackURL  string
}

func (p *Provider) Name() string {
	return p.providerName
}

func (p *Provider) SetName(name string) {
	p.providerName = name
}

func (p *Provider) GetSAMLServiceProvider() (*saml2.SAMLServiceProvider, error) {
	metadata, err := p.getIDPMetadata()
	if err != nil {
		return nil, err
	}

	certStore, err := p.getIDPCert(metadata)
	if err != nil {
		return nil, err
	}

	return &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      metadata.IDPSSODescriptor.SingleSignOnServices[0].Location,
		IdentityProviderIssuer:      metadata.EntityID,
		ServiceProviderIssuer:       SPIssuer,
		AssertionConsumerServiceURL: CallbackURL,
		SignAuthnRequests:           false,
		AudienceURI:                 SPAudienceURI,
		IDPCertificateStore:         &certStore,
		//SPKeyStore:                  randomKeyStore,
	}, nil
}

func (p *Provider) BeginAuth(state string) (goth.Session, error) {
	sp, err := p.GetSAMLServiceProvider()
	if err != nil {
		return &Session{}, err
	}

	stateValues := domain.GetSubPartKeyValues(state, "-", "=")

	relayState := ""
	returnTo := SPReturnTo

	if tempReturn, ok := stateValues["ReturnTo"]; ok {
		returnTo = tempReturn
	}

	clientID, _ := stateValues["ClientID"]
	if len(clientID) > 0 {
		returnTo += "?ClientID=" + clientID
	}

	relayState = "ReturnTo=" + returnTo

	authURL, err := sp.BuildAuthURL(relayState)

	if err != nil {
		return &Session{}, err
	}

	session := Session{
		AuthURL: authURL,
	}
	return &session, nil
}

// UnmarshalSession will unmarshal a JSON string into a session.
func (p *Provider) UnmarshalSession(data string) (goth.Session, error) {
	sess := &Session{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(sess)
	return sess, err
}

func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
	sess := session.(*Session)

	user := goth.User{
		AccessToken: "??????",
		Provider:    p.Name(),
		ExpiresAt:   sess.ExpiresAt,
	}
	return user, nil
}

func (p *Provider) Debug(debug bool) {}

func (p *Provider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	return nil, errors.New("Refresh token is not provided")
}

func (p *Provider) RefreshTokenAvailable() bool {
	return false
}

func (p *Provider) getIDPMetadata() (*types.EntityDescriptor, error) {
	rawMetadata, err := ioutil.ReadFile(IDPMetadataFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading IDP metadata file. %v", err.Error())
	}

	metadata := &types.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling IDP metadata file contents. %v", err.Error())
	}

	return metadata, nil
}

func (p *Provider) getIDPCert(metadata *types.EntityDescriptor) (dsig.MemoryX509CertificateStore, error) {
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}

	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {

		for idx, xcert := range kd.KeyInfo.X509Data.X509Certificates {

			if xcert.Data == "" {
				return certStore, fmt.Errorf("Metadata certificate(%d) must not be empty", idx)
			}
			certData, err := base64.StdEncoding.DecodeString(xcert.Data)
			if err != nil {
				return certStore, fmt.Errorf("Error getting IDP cert data from metadata. %v", err.Error())
			}

			idpCert, err := x509.ParseCertificate(certData)
			if err != nil {
				return certStore, fmt.Errorf("Error parsing IDP cert data from metadata. %v", err.Error())
			}

			certStore.Roots = append(certStore.Roots, idpCert)
		}
	}

	return certStore, nil
}

func GetSAMLAttributeFirstValue(attrName string, attributes []types.Attribute) string {
	for _, attr := range attributes {
		if attr.Name != attrName {
			continue
		}

		if len(attr.Values) > 0 {
			return attr.Values[0].Value
		}
		return ""
	}
	return ""
}
