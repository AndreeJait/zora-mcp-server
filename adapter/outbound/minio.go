package outbound

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/AndreeJait/zora-mcp-server/config"
	portOutbound "github.com/AndreeJait/zora-mcp-server/port/outbound"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage implements portOutbound.Storage using MinIO SDK.
type MinIOStorage struct {
	client *minio.Client
}

var _ portOutbound.Storage = (*MinIOStorage)(nil)

func NewMinIOStorage(cfg *config.AppConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
		Region: cfg.MinIO.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connect: %w", err)
	}

	return &MinIOStorage{client: client}, nil
}

func (m *MinIOStorage) GetObject(ctx context.Context, bucket, objectKey string) ([]byte, error) {
	obj, err := m.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj); err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}

	return buf.Bytes(), nil
}

func (m *MinIOStorage) GetPresignedURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned url: %w", err)
	}
	return url.String(), nil
}

func (m *MinIOStorage) Upload(ctx context.Context, bucket, objectKey string, reader io.Reader, objectSize int64, contentType string) error {
	opts := minio.PutObjectOptions{ContentType: contentType}
	if _, err := m.client.PutObject(ctx, bucket, objectKey, reader, objectSize, opts); err != nil {
		return fmt.Errorf("upload object: %w", err)
	}
	return nil
}

func (m *MinIOStorage) Delete(ctx context.Context, bucket, objectKey string) error {
	if err := m.client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}