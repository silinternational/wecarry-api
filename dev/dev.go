package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const AwsS3RegionEnv = "AWS_REGION"
const AwsS3EndpointEnv = "AWS_S3_ENDPOINT"
const AwsS3DisableSSLEnv = "AWS_S3_DISABLE_SSL"
const AwsS3BucketEnv = "AWS_S3_BUCKET"
const AwsS3AccessKeyIDEnv = "AWS_S3_ACCESS_KEY_ID"
const AwsS3SecretAccessKeyEnv = "AWS_S3_SECRET_ACCESS_KEY"

type awsConfig struct {
	awsAccessKeyID     string
	awsSecretAccessKey string
	awsEndpoint        string
	awsRegion          string
	awsS3Bucket        string
	awsDisableSSL      bool
	getPresignedUrl    bool
}

func getS3ConfigFromEnv() awsConfig {
	var a awsConfig
	a.awsAccessKeyID = os.Getenv(AwsS3AccessKeyIDEnv)
	a.awsSecretAccessKey = os.Getenv(AwsS3SecretAccessKeyEnv)
	a.awsEndpoint = os.Getenv(AwsS3EndpointEnv)
	a.awsRegion = os.Getenv(AwsS3RegionEnv)
	a.awsS3Bucket = os.Getenv(AwsS3BucketEnv)

	if disableSSL, err := strconv.ParseBool(os.Getenv(AwsS3DisableSSLEnv)); err == nil {
		a.awsDisableSSL = disableSSL
	}

	if len(a.awsEndpoint) > 0 {
		// a non-empty endpoint means minIO is in use, which doesn't support the S3 object URL scheme
		a.getPresignedUrl = true
	}
	return a
}

func createS3Service(config awsConfig) (*s3.S3, error) {
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

// createS3Bucket creates an S3 bucket with a name defined by an environment variable. If the bucket already
// exists, it will not return an error.
func createS3Bucket() error {
	config := getS3ConfigFromEnv()

	svc, err := createS3Service(config)
	if err != nil {
		return err
	}

	bucketName := os.Getenv(AwsS3BucketEnv)
	c := &s3.CreateBucketInput{Bucket: &bucketName}
	if _, err := svc.CreateBucket(c); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
			case s3.ErrCodeBucketAlreadyOwnedByYou:
			default:
				return err
			}
		}
	}
	return nil
}

// This utility establishes things in the dev environment that buffalo nor the app can set up on its own.
func main() {
	if err := createS3Bucket(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
