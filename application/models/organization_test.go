package models

import (
	"github.com/gofrs/uuid"
	"testing"
)

func TestFindOrgByUUID(t *testing.T) {
	_ = BounceTestDB()

	orgUuidStr := "51b5321d-2769-48a0-908a-7af1d15083e2"
	orgName := "ACME"

	// Load Organization test fixtures
	orgUuid1, _ := uuid.FromString(orgUuidStr)
	orgFix := Organizations{
		{
			ID:         1,
			Name:       orgName,
			Uuid:       orgUuid1,
			AuthType:   "saml2",
			AuthConfig: "[]",
		},
	}
	if err := CreateOrgs(orgFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

	org, err := FindOrgByUUID(orgUuidStr)
	if err != nil {
		t.Errorf("FindOrgByUUID() returned an error: %v", err)
	}
	if org.Name != orgName {
		t.Errorf("expected %v, found %v", orgName, org.Name)
	}
}
