package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

func LoadS3ConfigFromEnv() (S3Config, error) {
	cfg := S3Config{
		Endpoint:  strings.TrimSpace(os.Getenv("S3_ENDPOINT")),
		Region:    strings.TrimSpace(os.Getenv("S3_REGION")),
		Bucket:    strings.TrimSpace(os.Getenv("S3_BUCKET")),
		AccessKey: strings.TrimSpace(os.Getenv("S3_ACCESS_KEY")),
		SecretKey: strings.TrimSpace(os.Getenv("S3_SECRET_KEY")),
	}
	useSSL := strings.TrimSpace(os.Getenv("S3_USE_SSL"))
	if useSSL == "" {
		cfg.UseSSL = false
	} else {
		b, err := strconv.ParseBool(useSSL)
		if err != nil {
			return S3Config{}, fmt.Errorf("invalid S3_USE_SSL: %w", err)
		}
		cfg.UseSSL = b
	}

	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return S3Config{}, errors.New("missing required S3 env: S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY")
	}
	// Region can be empty for MinIO.
	return cfg, nil
}

type S3Storage struct {
	client *minio.Client
	bucket string
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	cl, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}

	return &S3Storage{client: cl, bucket: cfg.Bucket}, nil
}

type ObjectStat struct {
	ETag         string
	Size         int64
	ContentType  string
	LastModified time.Time
}

func (s *S3Storage) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) (ObjectStat, error) {
	info, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return ObjectStat{}, err
	}
	// minio-go returns ETag without quotes typically.
	return ObjectStat{ETag: info.ETag, Size: info.Size, ContentType: contentType, LastModified: time.Now().UTC()}, nil
}

func (s *S3Storage) GetObject(ctx context.Context, key string) (*minio.Object, ObjectStat, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, ObjectStat{}, err
	}
	st, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, ObjectStat{}, err
	}
	return obj, ObjectStat{ETag: st.ETag, Size: st.Size, ContentType: st.ContentType, LastModified: st.LastModified}, nil
}

func (s *S3Storage) StatObject(ctx context.Context, key string) (ObjectStat, error) {
	st, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return ObjectStat{}, err
	}
	return ObjectStat{ETag: st.ETag, Size: st.Size, ContentType: st.ContentType, LastModified: st.LastModified}, nil
}

func (s *S3Storage) DeleteObject(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

// SafeJoinAvatarPath ensures we don't allow path traversal.
func SafeJoinAvatarPath(prefix string, key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("empty key")
	}
	// Disallow attempts to escape.
	if strings.Contains(key, "..") || strings.ContainsAny(key, "\\") {
		return "", errors.New("invalid key")
	}
	// Remove leading slashes.
	key = strings.TrimLeft(key, "/")
	if prefix != "" {
		prefix = strings.Trim(prefix, "/")
		key = prefix + "/" + key
	}
	// Validate URL-safe-ish (allow slashes). Just ensure it's a valid path segment sequence.
	if strings.Contains(key, "//") {
		key = strings.ReplaceAll(key, "//", "/")
	}
	// Basic url.PathEscape isn't desired for keys, but validate it's parseable.
	if _, err := url.Parse("https://example.com/" + key); err != nil {
		return "", errors.New("invalid key")
	}
	return key, nil
}
