package usecase

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/AndreeJait/zora-mcp-server/port/inbound/upload"
	"github.com/AndreeJait/zora-mcp-server/port/outbound"
	"github.com/google/uuid"
)

type uploadUseCase struct {
	storage outbound.Storage
}

var _ upload.UseCase = (*uploadUseCase)(nil)

func NewUploadUseCase(storage outbound.Storage) upload.UseCase {
	return &uploadUseCase{storage: storage}
}

func (uc *uploadUseCase) Upload(ctx context.Context, bucket string, prefix string, filename string, contentType string, data []byte) (*upload.UploadResult, error) {
	if bucket == "" {
		bucket = "zora-files"
	}
	if prefix == "" {
		prefix = "uploads"
	}

	objectKey := fmt.Sprintf("%s/%s-%s", prefix, uuid.New().String(), filepath.Base(filename))
	size := int64(len(data))
	reader := bytes.NewReader(data)

	if err := uc.storage.Upload(ctx, bucket, objectKey, reader, size, contentType); err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	url, err := uc.storage.GetPresignedURL(ctx, bucket, objectKey, 24*time.Hour)
	if err != nil {
		url = fmt.Sprintf("%s/%s", bucket, objectKey)
	}

	return &upload.UploadResult{
		Bucket:      bucket,
		ObjectKey:   objectKey,
		URL:         url,
		ContentType: contentType,
		Size:        size,
	}, nil
}

func (uc *uploadUseCase) Get(ctx context.Context, bucket string, objectKey string) (*upload.GetResult, error) {
	if bucket == "" {
		bucket = "zora-files"
	}

	url, err := uc.storage.GetPresignedURL(ctx, bucket, objectKey, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	return &upload.GetResult{
		ObjectKey: objectKey,
		Bucket:    bucket,
		URL:       url,
	}, nil
}

func (uc *uploadUseCase) Delete(ctx context.Context, bucket string, objectKey string) error {
	if bucket == "" {
		bucket = "zora-files"
	}

	if err := uc.storage.Delete(ctx, bucket, objectKey); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

func (uc *uploadUseCase) Put(ctx context.Context, bucket string, objectKey string, contentType string, data []byte) (*upload.PutResult, error) {
	if bucket == "" {
		bucket = "zora-files"
	}

	size := int64(len(data))
	reader := bytes.NewReader(data)

	if err := uc.storage.Upload(ctx, bucket, objectKey, reader, size, contentType); err != nil {
		return nil, fmt.Errorf("put file: %w", err)
	}

	url, err := uc.storage.GetPresignedURL(ctx, bucket, objectKey, 24*time.Hour)
	if err != nil {
		url = fmt.Sprintf("%s/%s", bucket, objectKey)
	}

	return &upload.PutResult{
		ObjectKey: objectKey,
		Bucket:    bucket,
		URL:       url,
	}, nil
}