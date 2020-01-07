package aws

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"jaytaylor.com/html2text"

	"github.com/silinternational/wecarry-api/domain"
)

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

func getS3ConfigFromEnv() awsConfig {
	var a awsConfig
	a.awsAccessKeyID = domain.Env.AwsS3AccessKeyID
	a.awsSecretAccessKey = domain.Env.AwsS3SecretAccessKey
	a.awsEndpoint = domain.Env.AwsS3Endpoint
	a.awsRegion = domain.Env.AwsRegion
	a.awsS3Bucket = domain.Env.AwsS3Bucket
	a.awsDisableSSL = domain.Env.AwsS3DisableSSL

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
	config := getS3ConfigFromEnv()

	svc, err := createS3Service(config)
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
	config := getS3ConfigFromEnv()

	svc, err := createS3Service(config)
	if err != nil {
		return ObjectUrl{}, err
	}

	return getObjectURL(config, svc, key)
}

// CreateS3Bucket creates an S3 bucket with a name defined by an environment variable. If the bucket already
// exists, it will not return an error.
func CreateS3Bucket() error {
	env := domain.Env.GoEnv
	if env != "test" && env != "development" {
		return errors.New("CreateS3Bucket should only be used in test and development")
	}

	config := getS3ConfigFromEnv()

	svc, err := createS3Service(config)
	if err != nil {
		return err
	}

	c := &s3.CreateBucketInput{Bucket: &domain.Env.AwsS3Bucket}
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

// SendEmail sends a message using SES
func SendEmail(to, from, subject, body string) error {
	svc, err := createSESService(getSESConfigFromEnv())
	if err != nil {
		return fmt.Errorf("SendEmail failed creating SES service, %s", err)
	}

	input := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{Data: rawEmail(to, from, subject, body)},
		Source:     aws.String(from),
	}

	result, err := svc.SendRawEmail(input)
	if err != nil {
		return fmt.Errorf("SendEmail failed using SES, %s", err)
	}

	domain.Logger.Printf("Message sent using SES, message ID: %s", *result.MessageId)
	return nil
}

func rawEmail(to, from, subject, body string) []byte {
	tbody, err := html2text.FromString(body)
	if err != nil {
		domain.Logger.Printf("error converting html email to plain text ... %s", err.Error())
		tbody = body
	}

	var sb strings.Builder
	sb.WriteString("From: " + from + "\n")
	sb.WriteString("To:" + to + "\n")
	sb.WriteString("Subject:" + subject + "\n")
	sb.WriteString("MIME-Version: 1.0\n")

	b := &bytes.Buffer{}
	alternativeWriter := multipart.NewWriter(b)
	sb.WriteString(`Content-Type: multipart/alternative; boundary="` + alternativeWriter.Boundary() + `"` + "\n\n")

	w, _ := alternativeWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/plain", `charset="utf/8"`},
		"Content-Transfer-Encoding": {"7bit"},
		"Content-Disposition":       {"inline"},
	})
	_, _ = fmt.Fprintf(w, tbody)

	relatedWriter := multipart.NewWriter(b)
	_, _ = alternativeWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type": {`multipart/related; boundary="` + relatedWriter.Boundary() + `"`},
	})

	w, _ = relatedWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/html", `charset="utf/8"`},
		"Content-Transfer-Encoding": {"7bit"},
		"Content-Disposition":       {"inline"},
	})
	_, _ = fmt.Fprintf(w, body)

	w, _ = relatedWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"image/png"},
		"Content-Disposition":       {"inline"},
		"Content-ID":                {"<logo>"},
		"Content-Transfer-Encoding": {"base64"},
	})
	logo := `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`
	_, _ = fmt.Fprintf(w, logo)

	_ = relatedWriter.Close()
	_ = alternativeWriter.Close()

	sb.WriteString(b.String())
	fmt.Printf(sb.String())
	return []byte(sb.String())
}

func getSESConfigFromEnv() awsConfig {
	return awsConfig{
		awsAccessKeyID:     domain.Env.AwsSESAccessKeyID,
		awsSecretAccessKey: domain.Env.AwsSESSecretAccessKey,
		awsRegion:          domain.Env.AwsRegion,
	}
}

func createSESService(config awsConfig) (*ses.SES, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.awsAccessKeyID, config.awsSecretAccessKey, ""),
		Region:      aws.String(config.awsRegion),
	})
	if err != nil {
		return nil, err
	}
	return ses.New(sess), nil
}
