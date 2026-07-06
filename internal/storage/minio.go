package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type MinIOObjectStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIOObjectStorage Create a new minio storage
func NewMinIOObjectStorage(ctx context.Context, cfg MinIOConfig) (*MinIOObjectStorage, error) {
	if err := checkMinIOConfig(&cfg); err != nil {
		return nil, err
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	store := &MinIOObjectStorage{
		client: client,
		bucket: cfg.Bucket,
	}

	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

// checkMinIOConfig check minio config if it is valid
func checkMinIOConfig(cfg *MinIOConfig) error {
	if cfg == nil {
		return errors.New("nil MinIOConfig")
	}
	switch {
	case cfg.Endpoint == "":
		return errors.New("minio endpoint required")
	case cfg.AccessKey == "":
		return errors.New("minio access key required")
	case cfg.SecretKey == "":
		return errors.New("minio secret key required")
	case cfg.Bucket == "":
		return errors.New("minio bucket required")
	}
	return nil
}

// ensureBucket ensure bucket exist
// return if exists, else create a new bucket
func (s *MinIOObjectStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check minio bucket %q: %w", s.bucket, err)
	}

	if exists {
		return nil
	}

	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create minio bucket %q: %w", s.bucket, err)
	}

	return nil
}

// Put Update the minio
func (s *MinIOObjectStorage) Put(ctx context.Context, key string, reader io.Reader, size int64, opts PutOptions) (PutResult, error) {
	key = normalizeObjectKey(key)
	if key == "" {
		return PutResult{}, errors.New("object key is empty")
	}

	info, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType:  opts.ContentType,
		UserMetadata: opts.Metadata,
	})
	if err != nil {
		return PutResult{}, fmt.Errorf("put minio object %q: %w", key, err)
	}

	return PutResult{
		Key:  key,
		ETag: info.ETag,
		Size: info.Size,
	}, nil
}

func (s *MinIOObjectStorage) Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	key = normalizeObjectKey(key)
	if key == "" {
		return nil, ObjectInfo{}, fmt.Errorf("object key is empty")
	}

	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, ObjectInfo{}, fmt.Errorf("get minio object %q: %w", key, err)
	}

	stat, err := obj.Stat()
	if err != nil {
		return nil, ObjectInfo{}, fmt.Errorf("stat minio object %q: %w", key, err)
	}

	return obj, ObjectInfo{
		Key:          key,
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		ETag:         stat.ETag,
		LastModified: stat.LastModified,
	}, nil
}

func (s *MinIOObjectStorage) Delete(ctx context.Context, key string) error {
	key = normalizeObjectKey(key)
	if key == "" {
		return fmt.Errorf("object key is empty")
	}

	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete minio object %q: %w", key, err)
	}

	return nil
}

func (s *MinIOObjectStorage) PresignedGet(ctx context.Context, key string, expire time.Duration) (string, error) {
	key = normalizeObjectKey(key)
	if key == "" {
		return "", fmt.Errorf("object key is empty")
	}

	if expire <= 0 {
		expire = time.Hour
	}

	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, expire, nil)
	if err != nil {
		return "", fmt.Errorf("presigned get minio object %q: %w", key, err)
	}

	return u.String(), nil
}
