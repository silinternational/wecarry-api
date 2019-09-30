package aws

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

const AwsS3RegionEnv = "AWS_REGION"
const AwsS3EndpointEnv = "AWS_S3_ENDPOINT"
const AwsS3DisableSSLEnv = "AWS_S3_DISABLE_SSL"
const AwsS3BucketEnv = "AWS_S3_BUCKET"
const AwsS3AccessKeyIDEnv = "AWS_S3_ACCESS_KEY_ID"
const AwsS3SecretAccessKeyEnv = "AWS_S3_SECRET_ACCESS_KEY"

type ObjectUrl struct {
	Url        string
	Expiration time.Time
}

type awsConfig struct {
	awsAccessKeyID     string
	awsSecretAccessKey string
	awsEndpoint        string
	awsRegion          string
	awsS3Bucket        string
	awsDisableSSL      bool
	getPresignedUrl    bool
}

// presigned URL expiration
const urlLifespan = 10 * time.Minute

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

func GetS3ConfigFromEnv() awsConfig {
	var a awsConfig
	a.awsAccessKeyID = envy.Get(AwsS3AccessKeyIDEnv, "")
	a.awsSecretAccessKey = envy.Get(AwsS3SecretAccessKeyEnv, "")
	a.awsEndpoint = envy.Get(AwsS3EndpointEnv, "")
	a.awsRegion = envy.Get(AwsS3RegionEnv, "")
	a.awsS3Bucket = envy.Get(AwsS3BucketEnv, "")

	if disableSSL, err := strconv.ParseBool(envy.Get(AwsS3DisableSSLEnv, "false")); err == nil {
		a.awsDisableSSL = disableSSL
	}

	if len(a.awsEndpoint) > 0 {
		// a non-empty endpoint means minIO is in use, which doesn't support the S3 object URL scheme
		a.getPresignedUrl = true
	}
	return a
}

func CreateS3Service(config awsConfig) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.awsAccessKeyID, config.awsSecretAccessKey, ""),
		Endpoint:         aws.String(config.awsEndpoint),
		Region:           aws.String(config.awsRegion),
		DisableSSL:       aws.Bool(config.awsDisableSSL),
		S3ForcePathStyle: aws.Bool(len(config.awsEndpoint) > 0),
	})
	svc := s3.New(sess)

	return svc, err
}

func getObjectURL(config awsConfig, svc *s3.S3, key string) (ObjectUrl, error) {
	var objectUrl ObjectUrl

	if !config.getPresignedUrl {
		objectUrl.Url = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", config.awsS3Bucket, url.PathEscape(key))
		objectUrl.Expiration = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)
		return objectUrl, nil
	}

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(config.awsS3Bucket),
		Key:    aws.String(key),
	})

	if newUrl, err := req.Presign(urlLifespan); err == nil {
		objectUrl.Url = newUrl
		// return a time slightly before the actual url expiration to account for delays
		objectUrl.Expiration = time.Now().Add(urlLifespan - time.Minute)
	} else {
		return objectUrl, err
	}

	return objectUrl, nil
}

// StoreFile saves content in an AWS S3 bucket or compatible storage, depending on environment configuration.
func StoreFile(key, contentType string, content []byte) (ObjectUrl, error) {
	config := GetS3ConfigFromEnv()

	svc, err := CreateS3Service(config)
	if err != nil {
		return ObjectUrl{}, err
	}

	acl := ""
	if !config.getPresignedUrl {
		acl = "public-read"
	}
	if _, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(config.awsS3Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         aws.String(acl),
		Body:        bytes.NewReader(content),
	}); err != nil {
		return ObjectUrl{}, err
	}

	objectUrl, err := getObjectURL(config, svc, key)
	if err != nil {
		return ObjectUrl{}, err
	}

	return objectUrl, nil
}

// GetFileURL retrieves a URL from which a stored object can be loaded. The URL should not require external
// credentials to access. It may reference a file with public_read access or it may be a pre-signed URL.
func GetFileURL(key string) (ObjectUrl, error) {
	config := GetS3ConfigFromEnv()

	svc, err := CreateS3Service(config)
	if err != nil {
		return ObjectUrl{}, err
	}

	return getObjectURL(config, svc, key)
}
