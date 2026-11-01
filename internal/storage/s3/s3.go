package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"stitchforce-server/config"
	"stitchforce-server/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	client    *awsS3.Client
	presigner *awsS3.PresignClient
	bucket    string
}

func New(cfg config.StorageConfig) (*Storage, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, errors.New("storage endpoint is required")
	}

	if strings.TrimSpace(cfg.Region) == "" {
		return nil, errors.New("storage region is required")
	}

	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, errors.New("storage bucket is required")
	}

	if strings.TrimSpace(cfg.AccessKey) == "" ||
		strings.TrimSpace(cfg.SecretKey) == "" {
		return nil, errors.New("storage credentials are required")
	}

	awsConfig := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
	}

	client := awsS3.NewFromConfig(awsConfig, func(o *awsS3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = cfg.ForcePathStyle
	})

	return &Storage{
		client:    client,
		presigner: awsS3.NewPresignClient(client),
		bucket:    cfg.Bucket,
	}, nil
}

func (s *Storage) Upload(
	ctx context.Context,
	key string,
	file io.Reader,
	size int64,
	contentType string,
) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("storage key is required")
	}

	if file == nil {
		return errors.New("file is required")
	}

	if size < 0 {
		return errors.New("file size cannot be negative")
	}

	if strings.TrimSpace(contentType) == "" {
		return errors.New("content type is required")
	}

	_, err := s.client.PutObject(ctx, &awsS3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}

	return nil
}

func (s *Storage) Delete(
	ctx context.Context,
	key string,
) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("storage key is required")
	}

	_, err := s.client.DeleteObject(ctx, &awsS3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}

func (s *Storage) SignedURL(
	ctx context.Context,
	key string,
	expiry time.Duration,
) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("storage key is required")
	}

	if expiry <= 0 {
		return "", errors.New("expiry must be greater than zero")
	}

	result, err := s.presigner.PresignGetObject(
		ctx,
		&awsS3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
		awsS3.WithPresignExpires(expiry),
	)
	if err != nil {
		return "", fmt.Errorf("generate signed URL: %w", err)
	}

	return result.URL, nil
}

var _ storage.Storage = (*Storage)(nil)
