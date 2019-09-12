package domain

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gobuffalo/envy"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type ObjectUrl struct {
	Url        string
	Expiration time.Time
}

// presigned URL expiration is 7 days
const urlExpiration = 7 * 24 * time.Hour

func StoreFile(key, contentType string, content []byte, getPresignedUrl bool) (ObjectUrl, error) {
	awsAccessKeyID := envy.Get("AWS_ACCESS_KEY_ID", "")
	awsSecretAccessKey := envy.Get("AWS_SECRET_ACCESS_KEY", "")
	awsEndpoint := envy.Get("AWS_ENDPOINT", "")
	awsRegion := envy.Get("AWS_REGION", "")
	awsS3Bucket := envy.Get("AWS_S3_BUCKET", "")

	awsDisableSSL, err := strconv.ParseBool(envy.Get("AWS_DISABLE_SSL", "false"))
	if err != nil {
		awsDisableSSL = false
	}

	if len(awsEndpoint) > 0 {
		// a non-empty endpoint means minio is in use, which doesn't support the S3 object URL scheme
		getPresignedUrl = true
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
		Endpoint:         aws.String(awsEndpoint),
		Region:           aws.String(awsRegion),
		DisableSSL:       aws.Bool(awsDisableSSL),
		S3ForcePathStyle: aws.Bool(len(awsEndpoint) > 0),
	})
	if err != nil {
		return ObjectUrl{}, err
	}

	svc := s3.New(sess)

	r := bytes.NewReader(content)

	acl := ""
	if !getPresignedUrl {
		acl = "public-read"
	}
	if _, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(awsS3Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         aws.String(acl),
		Body:        r,
	}); err != nil {
		return ObjectUrl{}, err
	}

	var objectUrl ObjectUrl
	if getPresignedUrl {
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(awsS3Bucket),
			Key:    aws.String(key),
		})
		objectUrl.Url, err = req.Presign(urlExpiration)
		if err != nil {
			return ObjectUrl{}, err
		}

		// return a time slightly before the actual url expiration
		objectUrl.Expiration = time.Now().Add(urlExpiration - 1*time.Minute)
	} else {
		objectUrl.Url = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", awsS3Bucket, url.PathEscape(key))
		objectUrl.Expiration = time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC)
	}

	return objectUrl, nil
}
