package actions

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/silinternational/handcarry-api/auth/saml"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"

	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

const SAML2Provider = "saml2"
const SAMLResponseKey = "SAMLResponse"
const IDPMetadataFile = "./samlmetadata/idp-metadata.xml"

func AuthLogin(c buffalo.Context) error {

	clientID := c.Param("client_id")
	if clientID == "" {
		return fmt.Errorf("client_id is required to login")
	}
	c.Session().Set("ClientID", clientID)

	returnTo := c.Param("ReturnTo")
	if returnTo == "" {
		returnTo = "/"
	}
	c.Session().Set("ReturnTo", returnTo)

	err := c.Session().Save()
	if err != nil {
		return err
	}

	sp, err := getSAML2Provider()
	if err != nil {
		return err
	}

	authURL, err := sp.BuildAuthURL("")

	return c.Redirect(302, authURL)
}

func AuthCallback(c buffalo.Context) error {

	returnTo := envy.Get("UI_URL", "/")

	clientID := c.Session().Get("ClientID").(string)
	if clientID == "" {
		return fmt.Errorf("client_id is required to login")
	}

	samlResponse, err := samlResponse(c)
	if err != nil {
		return c.Error(401, err)
	}

	samlUser, err := getSamlUserFromAssertion(samlResponse)
	if err != nil {
		return c.Error(401, err)
	}

	// Get an existing User with the current auth org uid
	u := &models.User{}
	tx := c.Value("tx").(*pop.Connection)
	err = u.FindOrCreateFromSamlUser(tx, 1, samlUser)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, expiresAt, err := u.CreateAccessToken(tx, clientID)
	if err != nil {
		return errors.WithStack(err)
	}

	if err = tx.Save(u); err != nil {
		return errors.WithStack(err)
	}

	returnToURL := fmt.Sprintf("%s/?access_token=%s&expires=%v", returnTo, accessToken, expiresAt)

	return c.Redirect(302, returnToURL)
}

func AuthDestroy(c buffalo.Context) error {

	bearerToken := domain.GetBearerTokenFromRequest(c.Request())
	if bearerToken == "" {
		return errors.WithStack(fmt.Errorf("no Bearer token provided"))
	}

	user, err := models.FindUserByAccessToken(bearerToken)
	if err != nil {
		return errors.WithStack(err)
	}

	var logoutURL string

	if user.AuthOrg.AuthType == SAML2Provider {
		// TODO get logout url from user.AuthOrg.AuthConfig
		logoutURL = os.Getenv("SAML2_LOGOUT_URL")
		err := models.DeleteAccessToken(bearerToken)
		if err != nil {
			return err
		}
	}

	if logoutURL == "" {
		logoutURL = "/"
	}

	c.Session().Clear()
	return c.Redirect(302, logoutURL)
}

func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		bearerToken := domain.GetBearerTokenFromRequest(c.Request())
		if bearerToken == "" {
			return errors.WithStack(fmt.Errorf("no Bearer token provided"))
		}

		user, err := models.FindUserByAccessToken(bearerToken)
		if err != nil {
			return errors.WithStack(err)
		}
		c.Set("current_user", user)

		return next(c)
	}
}

func samlResponse(c buffalo.Context) (string, error) {
	reqData, err := domain.GetRequestData(c.Request())
	if err != nil {
		return "", err
	}

	samlResponse := domain.GetFirstStringFromSlice(reqData[SAMLResponseKey])
	if samlResponse == "" {
		return "", fmt.Errorf("%s not found in request", SAMLResponseKey)
	}

	return samlResponse, nil
}

func samlAttribute(attrName, samlResponse string) (string, error) {
	sp, err := getSAML2Provider()
	if err != nil {
		return "", err
	}

	response, err := sp.ValidateEncodedResponse(samlResponse)

	if err != nil {
		return "", fmt.Errorf("could not validate %s. %v", SAMLResponseKey, err.Error())
	}

	if response == nil {
		return "", fmt.Errorf("got nil response validating %s", SAMLResponseKey)
	}

	assertions := response.Assertions
	if len(assertions) < 1 {
		return "", fmt.Errorf("did not get any SAML assertions")
	}

	attrVal := getSAMLAttributeFirstValue(attrName, assertions[0].AttributeStatement.Attributes)
	if attrVal == "" {
		return "", fmt.Errorf("no value found for %s", attrName)
	}

	return attrVal, nil
}

func getIDPMetadata() (*types.EntityDescriptor, error) {
	rawMetadata, err := ioutil.ReadFile(IDPMetadataFile)
	if err != nil {
		return nil, fmt.Errorf("error reading IDP metadata file. %v", err.Error())
	}

	metadata := &types.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling IDP metadata file contents. %v", err.Error())
	}

	return metadata, nil
}

func getIDPCert(metadata *types.EntityDescriptor) (dsig.MemoryX509CertificateStore, error) {
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}

	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {

		for idx, xcert := range kd.KeyInfo.X509Data.X509Certificates {

			if xcert.Data == "" {
				return certStore, fmt.Errorf("metadata certificate(%d) must not be empty", idx)
			}
			certData, err := base64.StdEncoding.DecodeString(xcert.Data)
			if err != nil {
				return certStore, fmt.Errorf("error getting IDP cert data from metadata. %v", err.Error())
			}

			idpCert, err := x509.ParseCertificate(certData)
			if err != nil {
				return certStore, fmt.Errorf("error parsing IDP cert data from metadata. %v", err.Error())
			}

			certStore.Roots = append(certStore.Roots, idpCert)
		}
	}

	return certStore, nil
}

func getSAML2Provider() (*saml2.SAMLServiceProvider, error) {

	// metadata, err := getIDPMetadata()
	// if err != nil {
	// 	return &saml2.SAMLServiceProvider{}, err
	// }
	//
	// certStore, err := getIDPCert(metadata)
	// if err != nil {
	// 	return &saml2.SAMLServiceProvider{}, err
	// }

	idpSsoUrl := envy.Get("SAML_IDP_SSO_URL", "")
	if idpSsoUrl == "" {
		msg := "unable to get SAML_IDP_SSO_URL from environment"
		return &saml2.SAMLServiceProvider{}, fmt.Errorf(msg)
	}

	idpEntityId := envy.Get("SAML_IDP_ENTITY_ID", "")
	if idpSsoUrl == "" {
		msg := "unable to get SAML_IDP_ENTITY_ID from environment"
		return &saml2.SAMLServiceProvider{}, fmt.Errorf(msg)
	}

	idpCertData := envy.Get("SAML_IDP_CERT_DATA", "")
	if idpCertData == "" {
		msg := "unable to get SAML_IDP_CERT_DATA from environment"
		return &saml2.SAMLServiceProvider{}, fmt.Errorf(msg)
	}

	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}
	certData, err := base64.StdEncoding.DecodeString(idpCertData)
	if err != nil {
		return &saml2.SAMLServiceProvider{}, fmt.Errorf("error getting IDP cert data from metadata. %v", err.Error())
	}

	idpCert, err := x509.ParseCertificate(certData)
	if err != nil {
		return &saml2.SAMLServiceProvider{}, fmt.Errorf("error parsing IDP cert data from metadata. %v", err.Error())
	}

	certStore.Roots = append(certStore.Roots, idpCert)

	host := envy.Get("HOST", "")

	sp := &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      idpSsoUrl,
		IdentityProviderIssuer:      idpEntityId,
		ServiceProviderIssuer:       host,
		AssertionConsumerServiceURL: fmt.Sprintf("%s/%s", host, "auth/saml2/callback/"),
		SignAuthnRequests:           false,
		AudienceURI:                 host,
		IDPCertificateStore:         &certStore,
		//SPKeyStore:                  randomKeyStore,
	}

	return sp, nil
}

func getSAMLAttributeFirstValue(attrName string, attributes []types.Attribute) string {
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

func getSamlUserFromAssertion(assertion string) (saml.SamlUser, error) {
	firstName, err := samlAttribute("givenName", assertion)
	if err != nil {
		return saml.SamlUser{}, err
	}

	lastName, err := samlAttribute("sn", assertion)
	if err != nil {
		return saml.SamlUser{}, err
	}

	email, err := samlAttribute("mail", assertion)
	if err != nil {
		return saml.SamlUser{}, err
	}

	userID, err := samlAttribute("eduPersonTargetID", assertion)
	if err != nil {
		return saml.SamlUser{}, err
	}

	return saml.SamlUser{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		UserID:    userID,
	}, nil
}
