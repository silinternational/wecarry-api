package actions

import (
	"testing"
	"time"

	"github.com/silinternational/wecarry-api/aws"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type UserQueryFixtures struct {
	Users       models.Users
	CurrentUser models.User
	ClientID    string
	AccessToken string
}

func Fixtures_UserQuery(as *ActionSuite, t *testing.T) UserQueryFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := models.Users{
		{
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			Uuid:      domain.GetUuid(),
			AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin),
		},
		{
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
			Uuid:      domain.GetUuid(),
		},
	}

	for i := range users {
		if err := as.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user ... %v", err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := models.UserOrganizations{
		{
			OrganizationID: org.ID,
			UserID:         users[0].ID,
			AuthID:         "auth_user1",
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: org.ID,
			UserID:         users[1].ID,
			AuthID:         "auth_user2",
			AuthEmail:      users[1].Email,
		},
	}

	for i := range userOrgs {
		if err := as.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	clientID := "12345678"
	accessToken := "ABCDEFGHIJKLMONPQRSTUVWXYZ123456"
	hash := models.HashClientIdAccessToken(clientID + accessToken)

	userAccessToken := models.UserAccessToken{
		UserID:             users[0].ID,
		UserOrganizationID: userOrgs[0].ID,
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

	var f models.File

	if err := f.Store("photo.gif", []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	if _, err := users[1].AttachPhoto(f.UUID.String()); err != nil {
		t.Errorf("failed to attach photo to user, %s", err)
		t.FailNow()
	}

	return UserQueryFixtures{
		Users:       users,
		ClientID:    clientID,
		AccessToken: accessToken,
	}
}

type OrgFixtures struct {
	Users models.Users
	Orgs  models.Organizations
}

func Fixtures_CreateOrganization(as *ActionSuite, t *testing.T) OrgFixtures {

	// Array indexes for convenience in references
	const (
		SalesAdmin    = 0
		OrgMember     = 1
		OrgAdmin      = 2
		OtherOrgAdmin = 3
		Org1          = 0
		Org2          = 1
	)

	users := models.Users{
		{
			Email:     "sales_admin@domain.com",
			FirstName: "Sales",
			LastName:  "Admin",
			Nickname:  "sales_admin",
			AdminRole: nulls.NewString(domain.AdminRoleSalesAdmin),
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "member@domain.com",
			FirstName: "Org",
			LastName:  "Member",
			Nickname:  "org_member",
			AdminRole: nulls.String{},
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "admin@domain.com",
			FirstName: "Org",
			LastName:  "Admin",
			Nickname:  "org_admin",
			AdminRole: nulls.String{},
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "admin@other.com",
			FirstName: "Other Org",
			LastName:  "Admin",
			Nickname:  "other_org_admin",
			AdminRole: nulls.String{},
			Uuid:      domain.GetUuid(),
		},
	}
	for i := range users {
		err := as.DB.Create(&users[i])
		if err != nil {
			t.Errorf("unable to create user fixture %s: %s", users[i].Nickname, err)
		}
	}

	orgs := []models.Organization{
		{
			Name:       "Org1",
			Url:        nulls.String{},
			AuthType:   "saml2",
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
		{
			Name:       "Org2",
			Url:        nulls.String{},
			AuthType:   "saml2",
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
	}
	for i := range orgs {
		err := as.DB.Create(&orgs[i])
		if err != nil {
			t.Errorf("unable to create orgs fixture named %s: %s", orgs[i].Name, err)
		}
	}

	userOrgs := []models.UserOrganization{
		{
			OrganizationID: orgs[Org1].ID,
			UserID:         users[SalesAdmin].ID,
			Role:           models.UserOrganizationRoleMember,
			AuthID:         users[SalesAdmin].Nickname,
			AuthEmail:      users[SalesAdmin].Email,
		},
		{
			OrganizationID: orgs[Org1].ID,
			UserID:         users[OrgMember].ID,
			Role:           models.UserOrganizationRoleMember,
			AuthID:         users[OrgMember].Nickname,
			AuthEmail:      users[OrgMember].Email,
		},
		{
			OrganizationID: orgs[Org1].ID,
			UserID:         users[OrgAdmin].ID,
			Role:           models.UserOrganizationRoleAdmin,
			AuthID:         users[OrgAdmin].Nickname,
			AuthEmail:      users[OrgAdmin].Email,
		},
		{
			OrganizationID: orgs[Org2].ID,
			UserID:         users[OtherOrgAdmin].ID,
			Role:           models.UserOrganizationRoleAdmin,
			AuthID:         users[OtherOrgAdmin].Nickname,
			AuthEmail:      users[OtherOrgAdmin].Email,
		},
	}
	for i := range userOrgs {
		err := as.DB.Create(&userOrgs[i])
		if err != nil {
			t.Errorf("unable to create user orgs fixture for %s: %s", userOrgs[i].AuthID, err)
		}
	}

	accessTokenFixtures := []models.UserAccessToken{
		{
			UserID:             users[SalesAdmin].ID,
			UserOrganizationID: userOrgs[SalesAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(users[SalesAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OrgMember].ID,
			UserOrganizationID: userOrgs[OrgMember].ID,
			AccessToken:        models.HashClientIdAccessToken(users[OrgMember].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OrgAdmin].ID,
			UserOrganizationID: userOrgs[OrgAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(users[OrgAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OtherOrgAdmin].ID,
			UserOrganizationID: userOrgs[OtherOrgAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(users[OtherOrgAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
	}
	for i := range accessTokenFixtures {
		err := as.DB.Create(&accessTokenFixtures[i])
		if err != nil {
			t.Errorf("unable to create access token fixture for index %v: %s", i, err)
		}
	}

	return OrgFixtures{
		Users: users,
		Orgs:  orgs,
	}
}

type PostQueryFixtures struct {
	Posts       models.Posts
	Users       models.Users
	CurrentUser models.User
	ClientID    string
	AccessToken string
}

func Fixtures_PostQuery(as *ActionSuite, t *testing.T) PostQueryFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := models.Users{
		{
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			Uuid:      domain.GetUuid(),
			AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin),
		},
		{
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
			Uuid:      domain.GetUuid(),
		},
	}

	for i := range users {
		if err := as.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user ... %v", err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := models.UserOrganizations{
		{
			OrganizationID: org.ID,
			UserID:         users[0].ID,
			AuthID:         "auth_user1",
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: org.ID,
			UserID:         users[1].ID,
			AuthID:         "auth_user2",
			AuthEmail:      users[1].Email,
		},
	}

	for i := range userOrgs {
		if err := as.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	clientID := "12345678"
	accessToken := "ABCDEFGHIJKLMONPQRSTUVWXYZ123456"
	hash := models.HashClientIdAccessToken(clientID + accessToken)

	userAccessToken := models.UserAccessToken{
		UserID:             users[0].ID,
		UserOrganizationID: userOrgs[0].ID,
		AccessToken:        hash,
		ExpiresAt:          time.Now().Add(time.Hour),
	}

	if err := as.DB.Create(&userAccessToken); err != nil {
		t.Errorf("could not create test userAccessToken ... %v", err)
		t.FailNow()
	}

	// Load Post test fixtures
	posts := models.Posts{
		{
			CreatedByID:    users[0].ID,
			Type:           "Request",
			OrganizationID: org.ID,
			Title:          "A Request",
			Size:           "Small",
			Status:         "New",
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
		},
		{
			CreatedByID:    users[0].ID,
			Type:           "Offer",
			OrganizationID: org.ID,
			Title:          "An Offer",
			Size:           "Large",
			Status:         "New",
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[1].ID),
		},
	}

	for i := range posts {
		if err := as.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	var f models.File

	// attach photo
	if err := f.Store("photo.gif", []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	if _, err := posts[1].AttachPhoto(f.UUID.String()); err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
		t.FailNow()
	}

	// attach file
	if err := f.Store("dummy.pdf", []byte("%PDF-")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	if _, err := posts[1].AttachFile(f.UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
		t.FailNow()
	}

	return PostQueryFixtures{
		Posts:       posts,
		Users:       users,
		ClientID:    clientID,
		AccessToken: accessToken,
	}
}
