package imaged

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"time"
)

// Storage provides an interface for uploading and downloading files.
type Storage struct {
	svc        *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader

	Bucket string
}

// NewStorage creates a new storage object for a particular S3 bucket.
//
// The AWS credentials will be pulled from the environment.
func NewStorage(bucket string) (*Storage, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return &Storage{
		svc:        s3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		Bucket:     bucket,
	}, nil
}

// UploadBytes uploads a byte array to S3.
func (s *Storage) UploadBytes(ctx context.Context, key string, b []byte) (string, error) {
	reader := bytes.NewReader(b)
	input := &s3manager.UploadInput{
		Bucket: &s.Bucket,
		Key:    &key,
		Body:   reader,
	}
	result, err := s.uploader.UploadWithContext(ctx, input)
	if err != nil {
		return "", err
	}
	return result.Location, nil
}

// DownloadBytes downloads a byte array from S3.
func (s *Storage) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	buffer := aws.NewWriteAtBuffer(nil)
	input := &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	}
	_, err := s.downloader.DownloadWithContext(ctx, buffer, input)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// PublicURL generates a publicly-accessible URL for a file stored in S3.
func (s *Storage) PublicURL(ctx context.Context, key string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	}
	req, _ := s.svc.GetObjectRequest(input)
	return req.Presign(time.Hour)
}
