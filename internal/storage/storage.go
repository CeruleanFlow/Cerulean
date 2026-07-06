package storage

import (
	"context"
	"io"
	"strings"
	"time"
)

type PutOptions struct {
	ContentType string
	Metadata    map[string]string
}

type PutResult struct {
	Key  string
	ETag string
	Size int64
}

type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

type ObjectStorage interface {
	Put(ctx context.Context, key string, reader io.Reader, size int64, opts PutOptions) (PutResult, error)
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error)
	Delete(ctx context.Context, key string) error
	PresignedGet(ctx context.Context, key string, expire time.Duration) (string, error)
}

func normalizeObjectKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.TrimPrefix(key, "/")
	return key
}
