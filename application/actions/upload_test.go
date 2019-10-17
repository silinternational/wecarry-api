package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/gobuffalo/buffalo/binding"

	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/httptest"
	"github.com/silinternational/wecarry-api/models"
)

// UploadFixtures is for returning fixtures from Fixtures_Upload
type UploadFixtures struct {
	ClientID    string
	AccessToken string
}

// Fixtures_Upload creates fixtures for the Test_Upload test
func Fixtures_Upload(as *ActionSuite, t *testing.T) UploadFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	// Load User test fixture
	user := models.User{
		Uuid: domain.GetUuid(),
	}

	if err := as.DB.Create(&user); err != nil {
		t.Errorf("could not create test user ... %v", err)
		t.FailNow()
	}

	// Load UserOrganization test fixture
	userOrg := models.UserOrganization{
		OrganizationID: org.ID,
		UserID:         user.ID,
	}

	if err := as.DB.Create(&userOrg); err != nil {
		t.Errorf("could not create test user org ... %v", err)
		t.FailNow()
	}

	clientID := "12345678"
	accessToken := "ABCDEFGHIJKLMONPQRSTUVWXYZ123456"
	hash := models.HashClientIdAccessToken(clientID + accessToken)

	userAccessToken := models.UserAccessToken{
		UserID:             user.ID,
		UserOrganizationID: userOrg.ID,
		AccessToken:        hash,
		ExpiresAt:          time.Now().Add(time.Hour),
	}

	if err := as.DB.Create(&userAccessToken); err != nil {
		t.Errorf("could not create test userAccessToken ... %v", err)
		t.FailNow()
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	return UploadFixtures{
		ClientID:    clientID,
		AccessToken: accessToken,
	}
}

// Test_Upload tests the actions.UploadHandler function
func (as *ActionSuite) Test_Upload() {
	t := as.T()
	fixtures := Fixtures_Upload(as, t)

	type meta struct {
		File binding.File
	}

	const filename = "test.gif"

	f := httptest.File{
		ParamName: FileFieldName,
		FileName:  filename,
		Reader:    bytes.NewReader([]byte("GIF87a")),
	}

	req := as.HTML("/upload")
	req.Headers = map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", fixtures.ClientID+fixtures.AccessToken),
	}
	resp, err := req.MultiPartPost(&meta{}, f)
	as.NoError(err)
	as.Equal(200, resp.Code, "bad response code, body: \n%s", resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	var u UploadResponse
	err = json.Unmarshal(body, &u)
	if err != nil {
		t.Error(err)
	}

	as.Equal(filename, u.Name)
	as.NotEqual(domain.EmptyUUID, u.UUID)
	as.Regexp("^https?", u.URL)
	as.Equal("image/gif", u.ContentType)
	as.Equal(6, u.Size)
}
