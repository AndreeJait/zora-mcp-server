package echo

import (
	"io"
	"net/http"

	"github.com/AndreeJait/go-utility/v2/httpw/echow"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/upload"
	"github.com/labstack/echo/v5"
)

// RegisterUploadRoutes registers the file storage routes.
//
// @Summary      Upload a file to object storage
// @Description  Upload a file to MinIO and return the object key and presigned URL
// @Tags         storage
// @Accept       multipart/form-data
// @Produce      json
// @Param        file    formData  file    true  "File to upload"
// @Param        bucket  formData  string  false "Target bucket (default: zora-files)"
// @Param        prefix  formData  string  false "Path prefix (default: uploads)"
// @Success      200  {object}  upload.UploadResult
// @Failure      400  {object}  responsew.BaseResponse
// @Failure      500  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/upload [post]
func RegisterUploadRoutes(r RouteRegistrar, uploadUC upload.UseCase) {
	r.POST("/api/v1/upload", echow.Bind(handleUpload(uploadUC)))
}

func handleUpload(uploadUC upload.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		file, err := (*c).FormFile("file")
		if err != nil {
			return nil, err
		}

		src, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()

		data := make([]byte, file.Size)
		if _, err := src.Read(data); err != nil {
			return nil, err
		}

		bucket := (*c).FormValue("bucket")
		prefix := (*c).FormValue("prefix")

		result, err := uploadUC.Upload(c.Request().Context(), bucket, prefix, file.Filename, file.Header.Get("Content-Type"), data)
		if err != nil {
			return nil, err
		}

		return responsew.Success(result, "File uploaded"), nil
	}
}

// RegisterStorageRoutes registers the file retrieval and deletion routes.
func RegisterStorageRoutes(r RouteRegistrar, uploadUC upload.UseCase) {
	r.GET("/api/v1/storage/files/*", echow.Bind(handleGetFile(uploadUC)))
	r.PUT("/api/v1/storage/files/*", echow.Bind(handlePutFile(uploadUC)))
	r.DELETE("/api/v1/storage/files/*", echow.Bind(handleDeleteFile(uploadUC)))
}

// @Summary      Get a file by object key
// @Description  Get a presigned URL for a file in object storage by its object key
// @Tags         storage
// @Produce      json
// @Param        object_key  path  string  true  "Object key (e.g. uploads/abc-123-file.png)"
// @Param        bucket      query string  false "Bucket name (default: zora-files)"
// @Success      200  {object}  upload.GetResult
// @Failure      404  {object}  responsew.BaseResponse
// @Failure      500  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/storage/files/{object_key} [get]
func handleGetFile(uploadUC upload.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		// Echo strips the leading "/" from wildcard params, so we re-add it
		objectKey := "/" + (*c).Param("*")
		bucket := (*c).QueryParam("bucket")

		result, err := uploadUC.Get(c.Request().Context(), bucket, objectKey)
		if err != nil {
			return nil, err
		}

		return responsew.Success(result, "File retrieved"), nil
	}
}

// @Summary      Overwrite a file by object key
// @Description  Write (overwrite) content to a file in object storage by its object key
// @Tags         storage
// @Accept       application/octet-stream
// @Produce      json
// @Param        object_key    path  string  true  "Object key (e.g. uploads/abc-123-file.md)"
// @Param        bucket       query string  false "Bucket name (default: zora-files)"
// @Param        content_type query string  false "Content-Type of the file (default: application/octet-stream)"
// @Param        body         body  []byte  true  "File content (raw body)"
// @Success      200  {object}  upload.PutResult
// @Failure      400  {object}  responsew.BaseResponse
// @Failure      500  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/storage/files/{object_key} [put]
func handlePutFile(uploadUC upload.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		objectKey := "/" + (*c).Param("*")
		bucket := (*c).QueryParam("bucket")
		contentType := (*c).QueryParam("content_type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		data, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return nil, err
		}

		result, err := uploadUC.Put(c.Request().Context(), bucket, objectKey, contentType, data)
		if err != nil {
			return nil, err
		}

		return responsew.Success(result, "File updated"), nil
	}
}

// @Summary      Delete a file by object key
// @Description  Delete a file from object storage by its object key
// @Tags         storage
// @Produce      json
// @Param        object_key  path  string  true  "Object key (e.g. uploads/abc-123-file.png)"
// @Param        bucket      query string  false "Bucket name (default: zora-files)"
// @Success      200  {object}  responsew.BaseResponse
// @Failure      404  {object}  responsew.BaseResponse
// @Failure      500  {object}  responsew.BaseResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/storage/files/{object_key} [delete]
func handleDeleteFile(uploadUC upload.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		// Echo strips the leading "/" from wildcard params, so we re-add it
		objectKey := "/" + (*c).Param("*")
		bucket := (*c).QueryParam("bucket")

		if err := uploadUC.Delete(c.Request().Context(), bucket, objectKey); err != nil {
			return nil, err
		}

		return responsew.Success(nil, "File deleted"), nil
	}
}

// Ensure http.StatusOK is used by the compiler.
var _ = http.StatusOK