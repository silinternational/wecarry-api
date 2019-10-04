package grifts

import (
	"github.com/markbates/grift/grift"
	"github.com/silinternational/wecarry-api/aws"
)

var _ = grift.Namespace("minio", func() {

	_ = grift.Desc("minio", "Create a bucket in minIO")
	_ = grift.Add("createBucket", func(c *grift.Context) error {
		return aws.CreateS3Bucket()
	})
})
