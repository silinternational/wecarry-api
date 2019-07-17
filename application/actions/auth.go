package actions

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gobuffalo/pop"

	dsig "github.com/russellhaering/goxmldsig"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"

	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

const SAML2Provider = "saml2"
const SAMLResponseKey = "SAMLResponse"
const SAMLUserIDKey = "eduPersonTargetID"
const IDPMetadataFile = "./samlmetadata/idp-metadata.xml"
const CallbackURL = "http://handcarry.local:3000/auth/saml2/callback/"
const SPIssuer = "http://handcarry.local:3000"
const SPAudienceURI = SPIssuer
const SPReturnTo = "handcarry.local:3000/"

func AuthLogin(c buffalo.Context) error {

	returnTo := c.Param("ReturnTo")
	if returnTo == "" {
		returnTo = "/"
	}
	c.Session().Set("ReturnTo", returnTo)

	clientID := c.Param("client_id")
	if clientID == "" {
		return fmt.Errorf("client_id is required to login")
	}
	c.Session().Set("ClientID", clientID)

	err := c.Session().Save()
	if err != nil {
		return err
	}

	sp, err := getSAML2Provider()

	authURL, err := sp.BuildAuthURL("")

	return c.Redirect(302, authURL)
}

func AuthCallback(c buffalo.Context) error {

	returnTo := c.Session().Get("ReturnTo")
	if returnTo == "" {
		returnTo = "/"
	}

	clientID := c.Session().Get("ClientID")
	if clientID == "" {
		return fmt.Errorf("client_id is required to login")
	}

	samlResponse, err := samlResponse(c)
	if err != nil {
		return c.Error(401, err)
	}

	authOrgUid, err := samlAttribute(SAMLUserIDKey, samlResponse)
	if err != nil {
		return c.Error(401, err)
	}

	// Get an existing User with the current auth org uid
	u := &models.User{}
	tx := c.Value("tx").(*pop.Connection)
	q := tx.Where("auth_org_uid = ?", authOrgUid)
	exists, err := q.Exists("users")
	if err != nil {
		return errors.WithStack(err)
	}

	if !exists {
		// TODO see if user with email already exists
		return c.Error(404, fmt.Errorf("user with org_uid %s not found", authOrgUid))
	}

	if err = q.First(u); err != nil {
		return errors.WithStack(err)
	}

	accessToken, expiresAt, err := u.CreateAccessToken(tx, fmt.Sprintf("%v", clientID))
	if err != nil {
		return errors.WithStack(err)
	}

	if err = tx.Save(u); err != nil {
		return errors.WithStack(err)
	}

	// returnTo := relayValues["ReturnTo"]
	// if returnTo == "" {
	// 	returnTo = "/"
	// }
	//
	returnToURL := fmt.Sprintf("%s?access_token=%s&expires=%v", returnTo, accessToken, expiresAt)

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

	metadata, err := getIDPMetadata()
	if err != nil {
		return &saml2.SAMLServiceProvider{}, err
	}

	certStore, err := getIDPCert(metadata)
	if err != nil {
		return &saml2.SAMLServiceProvider{}, err
	}

	sp := &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      metadata.IDPSSODescriptor.SingleSignOnServices[0].Location,
		IdentityProviderIssuer:      metadata.EntityID,
		ServiceProviderIssuer:       SPIssuer,
		AssertionConsumerServiceURL: CallbackURL,
		SignAuthnRequests:           false,
		AudienceURI:                 SPAudienceURI,
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
