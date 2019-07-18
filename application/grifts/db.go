package grifts

import (
	uuid2 "github.com/gofrs/uuid"
	"github.com/markbates/grift/grift"
	"github.com/silinternational/handcarry-api/models"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	_ = grift.Add("seed", func(c *grift.Context) error {

		fixtureOrgs := []*models.Organization{
			{
				ID:         1,
				Uuid:       getUuid(),
				Name:       "AppsDev",
				AuthType:   "SAML",
				AuthConfig: "{}",
			},
			{
				ID:         2,
				Uuid:       getUuid(),
				Name:       "Other",
				AuthType:   "SAML",
				AuthConfig: "{}",
			},
		}

		for _, org := range fixtureOrgs {
			err := models.DB.Create(org)
			if err != nil {
				return err
			}
		}

		return nil
	})

})

func getUuid() string {
	uuid, _ := uuid2.NewV4()
	return uuid.String()
}
