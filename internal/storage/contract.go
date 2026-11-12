package storage

import (
	"context"
	"io"
	"time"
)

type Storage interface {
	Upload(
		ctx context.Context,
		key string,
		file io.Reader,
		size int64,
		contentType string,
	) error

	Delete(
		ctx context.Context,
		key string,
	) error

	SignedURL(
		ctx context.Context,
		key string,
		expiry time.Duration,
	) (string, error)

}
