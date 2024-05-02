package utils

import (
	"mime/multipart"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

const SixDays = 6 * 24 * time.Hour

var AccessKeyID string
var SecretAccessKey string
var MyRegion string

var StorageClient *session.Session

func CreateAWSStorageClient() {
	AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	MyRegion = os.Getenv("AWS_REGION")
	newSession, err := session.NewSession(
		&aws.Config{
			Region: aws.String(MyRegion),
			Credentials: credentials.NewStaticCredentials(
				AccessKeyID,
				SecretAccessKey,
				"", // a token will be created when the session it's used.
			),
		})
	if err != nil {
		logrus.Fatal("failed to initialize cloud storage")
	}
	StorageClient = newSession
}

func GenerateSignedURL(filePath string) (string, error) {
	svc := s3.New(StorageClient)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(filePath),
	})

	url, urlErr := req.Presign(SixDays)
	return url, urlErr
}

func UploadImageToBucket(file multipart.File, handler *multipart.FileHeader) error {
	// Upload the file to S3
	svc := s3.New(StorageClient)

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(handler.Filename),
		Body:   file,
	})
	if err != nil {
		return err
	}
	return nil
}
