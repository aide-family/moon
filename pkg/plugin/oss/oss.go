package oss

import (
	"context"
	"io"
)

// OssClient oss 客户端
type (
	OssClient interface {
		// UploadFile 上传文件
		UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error

		// DownloadFile 下载文件
		DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error)
		// DeleteFile 删除文件
		DeleteFile(ctx context.Context, objectName string) error

		// GetFileUrl 获取文件访问链接
		GetFileUrl(ctx context.Context, objectName string) (string, error)
	}
)
