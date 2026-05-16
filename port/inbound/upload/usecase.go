package upload

import "context"

// UploadResult holds the result of a file upload.
type UploadResult struct {
	Bucket      string `json:"bucket"`
	ObjectKey   string `json:"object_key"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// GetResult holds the result of a file retrieval by object key.
type GetResult struct {
	ObjectKey string `json:"object_key"`
	Bucket    string `json:"bucket"`
	URL       string `json:"url"`
}

// PutResult holds the result of overwriting a file by object key.
type PutResult struct {
	ObjectKey string `json:"object_key"`
	Bucket    string `json:"bucket"`
	URL       string `json:"url"`
}

// UseCase defines the inbound port for file storage operations.
type UseCase interface {
	Upload(ctx context.Context, bucket string, prefix string, filename string, contentType string, data []byte) (*UploadResult, error)
	Get(ctx context.Context, bucket string, objectKey string) (*GetResult, error)
	Put(ctx context.Context, bucket string, objectKey string, contentType string, data []byte) (*PutResult, error)
	Delete(ctx context.Context, bucket string, objectKey string) error
}