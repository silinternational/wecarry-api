package grifts

import (
	"github.com/gobuffalo/grift/grift"
	"github.com/silinternational/wecarry-api/aws"
)

var _ = grift.Namespace("minio", func() {
	_ = grift.Desc("minio", "seed minIO")
	_ = grift.Add("seed", func(c *grift.Context) error {
		return aws.CreateS3Bucket()
	})
})
