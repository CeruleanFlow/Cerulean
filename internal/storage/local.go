package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalObjectStorage struct {
	root string
}

func NewLocalObjectStorage(root string) (*LocalObjectStorage, error) {
	if root == "" {
		root = ".var/objects"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &LocalObjectStorage{root: root}, nil
}

func (s *LocalObjectStorage) Put(ctx context.Context, key string, reader io.Reader, size int64, opts PutOptions) (PutResult, error) {
	_ = ctx
	_ = opts

	key = normalizeObjectKey(key)
	path := s.pathFor(key)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return PutResult{}, err
	}

	file, err := os.Create(path)
	if err != nil {
		return PutResult{}, err
	}
	defer file.Close()

	written, err := io.Copy(file, reader)
	if err != nil {
		return PutResult{}, err
	}

	return PutResult{
		Key:  key,
		Size: written,
	}, nil
}

func (s *LocalObjectStorage) Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	_ = ctx

	key = normalizeObjectKey(key)
	path := s.pathFor(key)

	file, err := os.Open(path)
	if err != nil {
		return nil, ObjectInfo{}, err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, ObjectInfo{}, err
	}

	return file, ObjectInfo{
		Key:          key,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
	}, nil
}

func (s *LocalObjectStorage) Delete(ctx context.Context, key string) error {
	_ = ctx

	key = normalizeObjectKey(key)
	path := s.pathFor(key)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (s *LocalObjectStorage) PresignedGet(ctx context.Context, key string, expire time.Duration) (string, error) {
	_ = ctx
	_ = expire

	key = normalizeObjectKey(key)

	return "file://" + url.PathEscape(key), nil
}

func (s *LocalObjectStorage) pathFor(key string) string {
	clean := filepath.Clean(strings.TrimPrefix(key, "/"))
	return filepath.Join(s.root, clean)
}

var _ ObjectStorage = (*LocalObjectStorage)(nil)

func checkLocalObjectStorage(root string) error {
	if root == "" {
		return fmt.Errorf("local storage root is empty")
	}
	return nil
}
